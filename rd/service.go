package rd

import (
	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/driver/consul"
	"github.com/baowk/dilu-rd/driver/etcd"
	"github.com/baowk/dilu-rd/models"

	"github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type RDClient interface {
	Register(s *config.RegisterNode) error
	Deregister()
	Watch(s *config.DiscoveryNode) error
	GetService(name string, clientIp string) (*models.ServiceNode, error)
}

func NewRDClient(cfg *config.Config, logger *zap.SugaredLogger) (client RDClient, err error) {
	if cfg.Driver == "etcd" {
		c := clientv3.Config{
			Endpoints:   cfg.Endpoints,
			DialTimeout: cfg.Timeout,
		}
		client, err = etcd.NewClient(&c, logger)
	} else if cfg.Driver == "consul" {
		c := api.Config{
			Address:  cfg.Endpoints[0],
			Scheme:   cfg.Scheme,
			WaitTime: cfg.Timeout,
		}
		client, err = consul.NewClient(&c, logger)
	}
	if err != nil {
		return
	}
	for _, rs := range cfg.Registers {
		err = client.Register(rs)
		if err != nil {
			return
		}
	}
	for _, ds := range cfg.Discoveries {
		if ds.Enable {
			err = client.Watch(ds)
			if err != nil {
				return
			}
		}
	}
	return
}
