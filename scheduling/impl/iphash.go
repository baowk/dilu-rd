package impl

import (
	"hash/fnv"
	"net"

	"github.com/baowk/dilu-rd/models"
)

// IpHashHandler IP哈希调度，同一客户端IP固定分配到同一节点，支持会话保持
type IpHashHandler struct{}

func NewIpHashHandler() *IpHashHandler {
	return &IpHashHandler{}
}

func (h *IpHashHandler) GetServiceNode(nodes []*models.ServiceNode, _ string, clientIP string) *models.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}

	ip := extractIP(clientIP)
	hash := hashIP(ip)
	n := len(nodes)

	// 从 hash 位置开始找第一个可用节点
	for i := 0; i < n; i++ {
		idx := (int(hash) + i) % n
		if nodes[idx].Enable() {
			return nodes[idx]
		}
	}
	return nil
}

func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func hashIP(ip string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(ip))
	return h.Sum32()
}
