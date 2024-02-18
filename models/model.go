package models

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
)

type Protocol string

const (
	HTTP Protocol = "http"
	GRPC Protocol = "grpc"

// HTTPS Protocol = "https"
)

type ServiceNode struct {
	Namespace string           //命名空间
	Id        string           //服务id
	Name      string           //服务名
	Addr      string           //服务地址
	Port      int              //端口
	Protocol  Protocol         //协议
	Weight    int              //权重
	CurConns  int              //当前连接数
	Interval  time.Duration    //检测间隔
	Timeout   time.Duration    //服务检测超时时间
	HealthUrl string           //健康检查地址
	Tags      []string         //标签
	failCnt   int              //失败次数
	FailLimit int              //失败次数限制，到达失败次数就会被禁用
	enable    bool             //是否启用
	grpc      *grpc.ClientConn //grpc连接
}

func (n *ServiceNode) Enable() bool {
	return n.enable
}

func (n *ServiceNode) SetEnable(enable bool) {
	n.enable = enable
}

func (n *ServiceNode) ClearFailCnt() {
	n.failCnt = 0
}

func (n *ServiceNode) IncrFailCnt() {
	n.failCnt++
	fmt.Println("service node fail limit*****************", n.failCnt, n.FailLimit)
	if n.failCnt > n.FailLimit {
		n.enable = false
	}
}

func (n *ServiceNode) GetFailCnt() int {
	return n.failCnt
}

func (n *ServiceNode) GetUrl() string {
	return fmt.Sprintf("%s://%s:%d", n.Protocol, n.Addr, n.Port)
}

// func (n *ServiceNode) GetHttpClient() (*http.Client, error) {
// 	//http.NewClient(fmt.Sprintf("%s://%s:%d", n.Protocol, n.Addr, n.Port))
// 	return nil, nil
// }

func (n *ServiceNode) GetGrpcConn() (conn *grpc.ClientConn, err error) {
	if n.grpc != nil {
		conn = n.grpc
	} else {
		conn, err = grpc.Dial(fmt.Sprintf("%s:%d", n.Addr, n.Port), grpc.WithInsecure())
		if err == nil {
			n.grpc = conn
		}
	}
	return
}

// func (n *ServiceNode) GetRpcConn() (*rpc.Client, error) {
// 	if n.Protocol == "tcp" {
// 		return rpc.Dial(n.Protocol, fmt.Sprintf("%s:%d", n.Addr, n.Port))
// 	} else {
// 		return rpc.DialHTTP(n.Protocol, fmt.Sprintf("%s:%d", n.Addr, n.Port))
// 	}
// }

func (n *ServiceNode) Close() {
	n.enable = false
	if n.grpc != nil {
		n.grpc.Close()
	}
}
