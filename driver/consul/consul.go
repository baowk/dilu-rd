package consul

import (
	"errors"
	"sync"
	"time"

	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/scheduling"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type ConsulClient struct {
	client             *api.Client
	rwmutex            sync.RWMutex
	registered         []*models.ServiceNode
	discovered         map[string][]*models.ServiceNode //已发现的服务
	logger             *zap.SugaredLogger
	schedulingHandlers map[string]scheduling.SchedulingHandler
}

func NewClient(cfg *api.Config, logger *zap.SugaredLogger) (*ConsulClient, error) {
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{
		client:             client,
		rwmutex:            sync.RWMutex{},
		discovered:         make(map[string][]*models.ServiceNode),
		registered:         make([]*models.ServiceNode, 0),
		logger:             logger,
		schedulingHandlers: make(map[string]scheduling.SchedulingHandler),
	}, nil
}

func (c *ConsulClient) Register(s *models.ServiceNode) error {
	meta := map[string]string{
		"protocol": string(s.Protocol),
	}
	r := &api.AgentServiceRegistration{
		Namespace: s.Namespace,
		ID:        s.Id,
		Name:      s.Name,
		Port:      s.Port,
		Address:   s.Addr,
		Tags:      s.Tags,
		Meta:      meta,
	}
	var check *api.AgentServiceCheck
	if s.HealthUrl != "" {
		if s.Timeout <= 0 {
			s.Timeout = time.Duration(10) * time.Second
		}
		if s.Interval <= 0 {
			s.Interval = time.Duration(10) * time.Second
		}
		check = &api.AgentServiceCheck{
			Timeout:  s.Timeout.String(),
			Interval: s.Interval.String(),
		}
		if s.Protocol == "http" {
			check.HTTP = s.HealthUrl
		} else if s.Protocol == "grpc" {
			check.GRPC = s.HealthUrl
		}
		r.Check = check
	}

	err := c.client.Agent().ServiceRegister(r)
	if err != nil {
		c.logger.Error("register err", zap.Error(err))
		return err
	}
	return nil
}

func (c *ConsulClient) Deregister(r *models.ServiceNode) error {
	return c.client.Agent().ServiceDeregister(r.Id)
}

func (c *ConsulClient) Watch(s *config.DiscoveryNode) error {
	var lastIndex uint64 = 0
	go func(s *config.DiscoveryNode) {
		for {
			entries, qmeta, err := c.client.Health().Service(s.Name, s.Tag, false, &api.QueryOptions{
				Namespace: s.Namespace,
				WaitIndex: lastIndex,
			})
			if err != nil {
				c.logger.Error("watch", zap.Error(err))
				time.Sleep(time.Second * 1)
				continue
			}
			//fmt.Println(entries, qmeta, err)
			lastIndex = qmeta.LastIndex
			c.logger.Debug("watch", zap.Any("entries", entries), zap.Any("qmeta", qmeta))
			for _, entry := range entries {
				flag := 0
				status := entry.Checks.AggregatedStatus()
				c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("entry.Service.ID--------", entry.Service.ID))
				switch status {
				case api.HealthPassing:
					flag = 1 // 正常
				//case api.HealthMaint, api.HealthCritical, api.HealthWarning:
				default:
					flag = 2 // 删除
				}
				func() {
					c.rwmutex.Lock()
					if flag == 1 { // 新增
						if vs, ok := c.discovered[s.Name]; ok {
							found := false
							for _, v := range vs {
								if v.Id == entry.Service.ID {
									c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("update---", entry.Service.ID))
									found = true
									v.Addr = entry.Service.Address
									v.Port = entry.Service.Port
									v.Protocol = models.Protocol(entry.Service.Meta["protocol"])
									v.Namespace = entry.Service.Namespace
									v.Tags = entry.Service.Tags
									v.SetEnable(true)
									v.ClearFailCnt()
									break
								}
							}
							if !found {
								c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("add reset---", entry.Service.ID))
								ds := c.entryToServiceNode(entry, s)
								vs = append(vs, ds)
								c.discovered[s.Name] = vs
							}
						} else {
							c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("add 1---", entry.Service.ID))
							ds := c.entryToServiceNode(entry, s)
							c.discovered[s.Name] = []*models.ServiceNode{ds}
						}
					} else {
						c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("entry.Service.ID-----del---", entry.Service.ID))
						if vs, ok := c.discovered[s.Name]; ok {
							for i, v := range vs {
								if v.Id == entry.Service.ID {
									v.Close()
									vs = append(vs[:i], vs[i+1:]...)
									break
								}
							}
							c.discovered[s.Name] = vs
						} else {
							c.logger.Debug("not found")
						}
					}
					defer c.rwmutex.Unlock()
				}()
			}
			time.Sleep(time.Second * time.Duration(s.RetryTime))
		}
	}(s)
	return nil
}

func (c *ConsulClient) GetService(name string, clientIp string) (*models.ServiceNode, error) {
	c.rwmutex.RLock()
	defer c.rwmutex.RUnlock()
	if rs, ok := c.discovered[name]; ok && len(rs) > 0 {
		if sh, ok := c.schedulingHandlers[name]; ok {
			return sh.GetServiceNode(rs, name), nil
		}
	}
	return nil, errors.New("no service")
}

func (c *ConsulClient) entryToServiceNode(entry *api.ServiceEntry, s *config.DiscoveryNode) *models.ServiceNode {
	n := models.ServiceNode{
		Id:        entry.Service.ID,
		Namespace: entry.Service.Namespace,
		Name:      entry.Service.Service,
		Tags:      entry.Service.Tags,
		Addr:      entry.Service.Address,
		Port:      entry.Service.Port,
		Protocol:  models.Protocol(entry.Service.Meta["protocol"]),
		FailLimit: s.FailLimit,
	}
	n.SetEnable(true)
	c.schedulingHandlers[n.Name] = scheduling.GetHandler(s.SchedulingAlgorithm, c.logger)
	return &n
}
