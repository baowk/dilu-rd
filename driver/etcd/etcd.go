package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/scheduling"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EtcdClient struct {
	client             *clientv3.Client
	rwmutex            sync.RWMutex
	registered         []*models.ServiceNode
	discovered         map[string][]*models.ServiceNode //已发现的服务
	logger             *zap.SugaredLogger
	schedulingHandlers map[string]scheduling.SchedulingHandler
}

func NewClient(cfg *clientv3.Config, logger *zap.SugaredLogger) (*EtcdClient, error) {
	client, err := clientv3.New(*cfg)
	if err != nil {
		return nil, err
	}
	return &EtcdClient{
		client:     client,
		rwmutex:    sync.RWMutex{},
		discovered: make(map[string][]*models.ServiceNode),
		registered: make([]*models.ServiceNode, 0),
		logger:     logger,
	}, nil
}

func (c *EtcdClient) Register(s *models.ServiceNode) error {
	var err error
	go func() {
		kv := clientv3.NewKV(c.client)
		lease := clientv3.NewLease(c.client)
		var curLeaseId clientv3.LeaseID = 0
		for {
			if curLeaseId == 0 {
				leaseResp, err := lease.Grant(context.TODO(), 10)
				if err != nil {
					c.logger.Error(err)
					return
				}

				key := s.Name + fmt.Sprintf("%d", leaseResp.ID)
				b, err := json.Marshal(s)
				if err != nil {
					c.logger.Error(err)
				}
				if _, err := kv.Put(context.TODO(), key, string(b), clientv3.WithLease(clientv3.LeaseID(leaseResp.ID))); err != nil {
					c.logger.Error(err)
					return
				}
				curLeaseId = clientv3.LeaseID(leaseResp.ID)
			} else {
				// 续约租约，如果租约已经过期将curLeaseId复位到0重新走创建租约的逻辑
				if _, err := lease.KeepAliveOnce(context.TODO(), curLeaseId); err == rpctypes.ErrLeaseNotFound {
					curLeaseId = 0
					continue
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
	return err
}

func (c *EtcdClient) Deregister(s *models.ServiceNode) error {
	return nil
}

func (c *EtcdClient) Watch(s *config.DiscoveryNode) error {
	go func() {
		watcher := clientv3.NewWatcher(c.client)
		// Watch 服务目录下的更新
		watchChan := watcher.Watch(context.TODO(), s.Name, clientv3.WithPrefix())
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				c.logger.Info(event)
				func() {
					c.rwmutex.Lock()
					switch event.Type {
					case mvccpb.PUT: //PUT事件，目录下有了新key
						var rs *models.ServiceNode
						err := json.Unmarshal(event.Kv.Value, rs)
						if err != nil {
							c.logger.Error(err)
						}
						if vs, ok := c.discovered[string(event.Kv.Key)]; ok {
							found := false
							for _, v := range vs {
								if v.Id == rs.Id {
									v.Addr = rs.Addr
									v.Port = rs.Port
									v.Tags = rs.Tags
									v.Weight = rs.Weight
									v.Namespace = rs.Namespace
									v.Protocol = rs.Protocol
									v.SetEnable(true)
									v.ClearFailCnt()
									found = true
									break
								}
							}
							if !found {
								rs.SetEnable(true)
								rs.ClearFailCnt()
								vs = append(vs, rs)
								c.discovered[string(event.Kv.Key)] = vs
							}
						} else {
							rs.SetEnable(true)
							rs.ClearFailCnt()
							c.discovered[string(event.Kv.Key)] = []*models.ServiceNode{rs}
						}
					case mvccpb.DELETE: //DELETE事件，目录中有key被删掉(Lease过期，key 也会被删掉)
						var rs *models.ServiceNode
						err := json.Unmarshal(event.Kv.Value, rs)
						if err != nil {
							c.logger.Error(err)
						}
						if vs, ok := c.discovered[string(event.Kv.Key)]; ok {
							for i, v := range vs {
								if v.Id == rs.Id {
									v.Close()
									vs = append(vs[:i], vs[i+1:]...)
									break
								}
							}
							c.discovered[string(event.Kv.Key)] = vs
						} else {
							c.logger.Info("not found")
						}
					}
					defer c.rwmutex.Unlock()
				}()
			}
		}
	}()
	return nil
}

func (c *EtcdClient) GetService(name string, clientIp string) (*models.ServiceNode, error) {
	c.rwmutex.RLock()
	defer c.rwmutex.RUnlock()
	if rs, ok := c.discovered[name]; ok && len(rs) > 0 {
		if sh, ok := c.schedulingHandlers[name]; ok {
			return sh.GetServiceNode(rs, name), nil
		}
	}
	return nil, errors.New("no service")
}
