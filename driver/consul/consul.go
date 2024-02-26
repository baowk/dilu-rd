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
	registered         []*config.RegisterNode
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
		registered:         make([]*config.RegisterNode, 0),
		logger:             logger,
		schedulingHandlers: make(map[string]scheduling.SchedulingHandler),
	}, nil
}

func (c *ConsulClient) Register(s *config.RegisterNode) error {
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
	if s.HealthCheck != "" {
		check = &api.AgentServiceCheck{
			Timeout:  s.Timeout.String(),
			Interval: s.Interval.String(),
		}
		if s.Protocol == "http" {
			check.HTTP = s.HealthCheck
		} else if s.Protocol == "grpc" {
			check.GRPC = s.HealthCheck
		}
		r.Check = check
	}

	err := c.client.Agent().ServiceRegister(r)
	if err != nil {
		c.logger.Error("register err", zap.Error(err))
		return err
	}
	c.registered = append(c.registered, s)
	return nil
}

func (c *ConsulClient) Deregister() {
	for _, r := range c.registered {
		c.client.Agent().ServiceDeregister(r.Id)
	}
}

func (c *ConsulClient) Watch(s *config.DiscoveryNode) error {
	var lastIndex uint64 = 0
	c.schedulingHandlers[s.Name] = scheduling.GetHandler(s.SchedulingAlgorithm, c.logger)
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
			lastIndex = qmeta.LastIndex
			c.logger.Debug("watch", zap.Any("entries", entries), zap.Any("qmeta", qmeta))
			for _, entry := range entries {
				status := entry.Checks.AggregatedStatus()
				c.logger.Debug("watch", zap.Any("-------status--", status), zap.Any("entry.Service.ID--------", entry.Service.ID))
				switch status {
				case api.HealthPassing:
					c.putServiceNode(s, entry)
				//case api.HealthMaint, api.HealthCritical, api.HealthWarning:
				default:
					c.delServiceNode(s, entry)
				}
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

func (c *ConsulClient) putServiceNode(s *config.DiscoveryNode, entry *api.ServiceEntry) {
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()
	if vs, ok := c.discovered[s.Name]; ok {
		found := false
		for _, v := range vs {
			if v.Id == entry.Service.ID {
				c.logger.Debug("watch", zap.Any("update---", entry.Service.ID))
				found = true
				v.Addr = entry.Service.Address
				v.Port = entry.Service.Port
				v.Protocol = entry.Service.Meta["protocol"]
				v.Namespace = entry.Service.Namespace
				v.Tags = entry.Service.Tags
				v.SetEnable(true)
				v.ClearFailCnt()
				break
			}
		}
		if !found {
			c.logger.Debug("watch", zap.Any("add reset---", entry.Service.ID))
			ds := c.entryToServiceNode(entry, s)
			vs = append(vs, ds)
			c.discovered[s.Name] = vs
		}
	} else {
		c.logger.Debug("watch", zap.Any("add 1---", entry.Service.ID))
		ds := c.entryToServiceNode(entry, s)
		c.discovered[s.Name] = []*models.ServiceNode{ds}
	}
}

func (c *ConsulClient) delServiceNode(s *config.DiscoveryNode, entry *api.ServiceEntry) {
	c.rwmutex.Lock()
	defer c.rwmutex.Unlock()
	c.logger.Debug("watch", zap.Any("entry.Service.ID-----del---", entry.Service.ID))
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

func (c *ConsulClient) entryToServiceNode(entry *api.ServiceEntry, s *config.DiscoveryNode) *models.ServiceNode {
	r := config.RegisterNode{
		Id:        entry.Service.ID,
		Namespace: entry.Service.Namespace,
		Name:      entry.Service.Service,
		Tags:      entry.Service.Tags,
		Addr:      entry.Service.Address,
		Port:      entry.Service.Port,
		Protocol:  entry.Service.Meta["protocol"],
		FailLimit: s.FailLimit,
	}
	n := models.ServiceNode{
		RegisterNode: r,
	}
	n.SetEnable(true)

	return &n
}
