package config

import (
	"fmt"
	"time"

	"github.com/baowk/dilu-rd/config"
	"github.com/baowk/dilu-rd/models"
)

var (
	consulAddr      = "172.19.167.104:8500"
	etcdAddr        = "172.19.167.104:2379"
	serverAddr      = "172.19.167.104"
	serverPort      = 5000
	GrpcPort        = 5001
	ServiceName     = "test-api"
	GrpcServiceName = "grpc-test-api"
	ConsulDriver    = "consul"
	EctdDriver      = "etcd"
	Driver          = EctdDriver

	// scheme      = "http"
)

func GetCenterAddr() string {
	if Driver == "consul" {
		return consulAddr
	} else if Driver == "etcd" {
		return etcdAddr
	}
	return consulAddr
}

func GetConfig() *config.Config {
	fmt.Println(Driver, GetCenterAddr())
	return &config.Config{
		Endpoints: []string{GetCenterAddr()},
		Scheme:    "http",
		Timeout:   time.Duration(10) * time.Second,
		Enable:    true,
		Driver:    Driver,
		ServiceNodes: []*models.ServiceNode{
			&models.ServiceNode{
				//Namespace: "dilu",
				Name:      ServiceName,
				Addr:      serverAddr,
				Port:      serverPort,
				Protocol:  "http",
				HealthUrl: fmt.Sprintf("http://%s:%d/api/health", serverAddr, serverPort),
				Tags:      []string{"dev"},
				Interval:  5 * time.Second,
				Weight:    100,
				Timeout:   10 * time.Second,
				Id:        fmt.Sprintf("%s:%d", serverAddr, serverPort),
			},
			&models.ServiceNode{
				//Namespace: "dilu",
				Name:      GrpcServiceName,
				Addr:      serverAddr,
				Port:      GrpcPort,
				Protocol:  "grpc",
				HealthUrl: fmt.Sprintf("%s:%d/Health", serverAddr, GrpcPort),
				Tags:      []string{"dev"},
				Interval:  5 * time.Second,
				Weight:    100,
				Timeout:   10 * time.Second,
				Id:        fmt.Sprintf("%s:%d", serverAddr, GrpcPort),
			},
		},
	}
}

func GetDisConfig() *config.Config {
	fmt.Println(Driver, GetCenterAddr())
	return &config.Config{
		Endpoints: []string{GetCenterAddr()},
		Scheme:    "http",
		Timeout:   time.Duration(10) * time.Second,
		Enable:    true,
		Driver:    Driver,
		Discoveries: []*config.DiscoveryNode{
			&config.DiscoveryNode{
				Enable: true,
				//SchedulingAlgorithm: "random",
				Name:      ServiceName,
				FailLimit: 3,
				//Namespace:           "dilu",
				//Tag:  "dev",
			},
			&config.DiscoveryNode{
				Enable: true,
				//SchedulingAlgorithm: "random",
				Name:      GrpcServiceName,
				FailLimit: 3,
				//Namespace:           "dilu",
				//Tag:  "dev",
			},
		},
	}
}
