package rd

import (
	"fmt"
	"time"

	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/driver/consul"
	"github.com/baowk/dilu-rd/driver/etcd"
	"github.com/baowk/dilu-rd/models"

	"github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type RDClient interface {
	Register(s *config.RegisterNode) error
	Deregister()
	Watch(s *config.DiscoveryNode) error
	GetService(name string, clientIp string) (*models.ServiceNode, error)
}

func NewRDClient(cfg *config.Config) (client RDClient, err error) {
	if cfg.Driver == "etcd" {
		c := clientv3.Config{
			Endpoints:   cfg.Endpoints,
			DialTimeout: cfg.Timeout,
		}
		client, err = etcd.NewClient(&c)
	} else if cfg.Driver == "consul" {
		c := api.Config{
			Address:  cfg.Endpoints[0],
			Scheme:   cfg.Scheme,
			WaitTime: cfg.Timeout,
		}
		client, err = consul.NewClient(&c)
	} else {
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
	if err != nil {
		return
	}
	for _, rs := range cfg.Registers {
		if rs.Addr == "" || rs.Port <= 0 || rs.Port > 65535 {
			return nil, fmt.Errorf("register node addr or port is invalid: addr=%s, port=%d", rs.Addr, rs.Port)
		}
		if rs.Protocol != "http" && rs.Protocol != "grpc" {
			return nil, fmt.Errorf("register node protocol is invalid: %s", rs.Protocol)
		}
		if rs.Name == "" {
			return nil, fmt.Errorf("register node name is empty")
		}
		if rs.Id == "" {
			rs.Id = fmt.Sprintf("%s:%d", rs.Addr, rs.Port)
		}
		if rs.FailLimit <= 0 {
			rs.FailLimit = 3
		}
		if rs.Interval <= 0 {
			rs.Interval = time.Duration(5 * time.Second)
		}
		if rs.Timeout <= 0 {
			rs.Timeout = time.Duration(10 * time.Second)
		}
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
