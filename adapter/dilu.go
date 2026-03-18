package adapter

import (
	"fmt"
	"time"

	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/rd"
)

// DiluRegistry 适配 dilu-core 的 ServiceRegistry 接口
// 用法：
//
//	cfg := adapter.DefaultConfig("consul", []string{"localhost:8500"})
//	registry, _ := adapter.NewDiluRegistry(cfg)
//	core.GetApp().SetRegistry(registry)
type DiluRegistry struct {
	client rd.RDClient
	cfg    *config.Config
}

// DefaultConfig 创建默认注册配置
func DefaultConfig(driver string, endpoints []string) *config.Config {
	return &config.Config{
		Enable:    true,
		Driver:    driver,
		Endpoints: endpoints,
		Scheme:    "http",
		Timeout:   10 * time.Second,
	}
}

// NewDiluRegistry 创建 dilu-core 适配的注册中心客户端
func NewDiluRegistry(cfg *config.Config) (*DiluRegistry, error) {
	return &DiluRegistry{cfg: cfg}, nil
}

// Register 实现 dilu-core ServiceRegistry 接口
func (d *DiluRegistry) Register(name, addr string, port int, healthURL string, tags []string) error {
	d.cfg.Registers = []*config.RegisterNode{
		{
			Name:        name,
			Addr:        addr,
			Port:        port,
			Protocol:    "http",
			HealthCheck: healthURL,
			Tags:        tags,
			Id:          fmt.Sprintf("%s:%d", addr, port),
			FailLimit:   3,
			Interval:    5 * time.Second,
			Timeout:     10 * time.Second,
		},
	}

	client, err := rd.NewRDClient(d.cfg)
	if err != nil {
		return err
	}
	d.client = client
	return nil
}

// Deregister 实现 dilu-core ServiceRegistry 接口
func (d *DiluRegistry) Deregister() {
	if d.client != nil {
		d.client.Deregister()
	}
}

// GetClient 获取底层 RDClient（可用于服务发现）
func (d *DiluRegistry) GetClient() rd.RDClient {
	return d.client
}
