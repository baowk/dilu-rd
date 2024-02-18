package config

import (
	"time"

	"github.com/baowk/dilu-rd/models"
)

type Config struct {
	Enable       bool
	Driver       string
	Endpoints    []string
	Scheme       string
	Timeout      time.Duration
	ServiceNodes []*models.ServiceNode
	Discoveries  []*DiscoveryNode
}

type DiscoveryNode struct {
	Enable              bool   //启用发现
	Namespace           string //命名空间
	Name                string //服务名
	Tag                 string //标签
	SchedulingAlgorithm string //调度算法
	FailLimit           int    //已发现服务最大失败数
	RetryTime           int    //重试时间间隔 秒
}
