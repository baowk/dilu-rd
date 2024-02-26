package config

import (
	"time"
)

type Config struct {
	Enable      bool             `mapstructure:"enable" json:"enable" yaml:"enable"`
	Driver      string           `mapstructure:"driver" json:"driver" yaml:"driver"`
	Endpoints   []string         `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints"`
	Scheme      string           `mapstructure:"scheme" json:"scheme" yaml:"scheme"`
	Timeout     time.Duration    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	Registers   []*RegisterNode  `mapstructure:"registers" json:"registers" yaml:"registers"`
	Discoveries []*DiscoveryNode `mapstructure:"discoveries" json:"discoveries" yaml:"discoveries"`
}

type RegisterNode struct {
	Namespace   string        `mapstructure:"namespace" json:"namespace" yaml:"namespace"`          //命名空间
	Id          string        `mapstructure:"id" json:"id" yaml:"id"`                               //服务id
	Name        string        `mapstructure:"name" json:"name" yaml:"name"`                         //服务名
	Addr        string        `mapstructure:"addr" json:"addr" yaml:"addr"`                         //服务地址
	Port        int           `mapstructure:"port" json:"port" yaml:"port"`                         //端口
	Protocol    string        `mapstructure:"protocol" json:"protocol" yaml:"protocol"`             //协议
	Weight      int           `mapstructure:"weight" json:"weight" yaml:"weight"`                   //权重
	Interval    time.Duration `mapstructure:"interval" json:"interval" yaml:"interval"`             //检测间隔
	Timeout     time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`                //服务检测超时时间
	HealthCheck string        `mapstructure:"health-check" json:"health-check" yaml:"health-check"` //健康检查地址
	Tags        []string      `mapstructure:"tags" json:"tags" yaml:"tags"`                         //标签
	FailLimit   int           `mapstructure:"fail-limit" json:"fail-limit" yaml:"fail-limit"`       //失败次数限制，到达失败次数就会被禁用
}

// func (e *RegisterNode) GetInterval() time.Duration {
// 	if e.Interval == 0 {
// 		return time.Second * 5
// 	}
// 	return e.Interval
// }

// func (e *RegisterNode) GetTimeout() time.Duration {
// 	if e.Timeout == 0 {
// 		return time.Second * 10
// 	}
// 	return e.Timeout
// }

// func (e *RegisterNode) GetId() string {
// 	if e.Id == "" {
// 		return fmt.Sprintf("%s:%d", e.Addr, e.Port)
// 	}
// 	return e.Id
// }

// func (e *RegisterNode) GetFailLimit() int {
// 	if e.FailLimit == 0 {
// 		return 3
// 	}
// 	return e.FailLimit
// }

type DiscoveryNode struct {
	Enable              bool   `mapstructure:"enable" json:"enable" yaml:"enable"`                                           //启用发现
	Namespace           string `mapstructure:"namespace" json:"namespace" yaml:"namespace"`                                  //命名空间
	Name                string `mapstructure:"name" json:"name" yaml:"name"`                                                 //服务名
	Tag                 string `mapstructure:"tag" json:"tag" yaml:"tag"`                                                    //标签
	SchedulingAlgorithm string `mapstructure:"scheduling-algorithm" json:"scheduling-algorithm" yaml:"scheduling-algorithm"` //调度算法
	FailLimit           int    `mapstructure:"fail-limit" json:"fail-limit" yaml:"fail-limit"`                               //已发现服务最大失败数
	RetryTime           int    `mapstructure:"retry-time" json:"retry-time" yaml:"retry-time"`                               //重试时间间隔 秒
}
