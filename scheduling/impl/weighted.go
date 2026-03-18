package impl

import (
	"sync"

	"github.com/baowk/dilu-rd/models"
)

// WeightedRobinHandler 加权轮询调度
// 使用平滑加权轮询算法（Nginx 风格），避免同一节点连续被选中
type WeightedRobinHandler struct {
	mu             sync.Mutex
	currentWeights map[string][]int // name -> 每个节点的当前权重
}

func NewWeightedRobinHandler() *WeightedRobinHandler {
	return &WeightedRobinHandler{
		currentWeights: make(map[string][]int),
	}
}

func (w *WeightedRobinHandler) GetServiceNode(nodes []*models.ServiceNode, name string, _ string) *models.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	n := len(nodes)

	// 初始化或长度变化时重置
	if cw, ok := w.currentWeights[name]; !ok || len(cw) != n {
		w.currentWeights[name] = make([]int, n)
	}
	cw := w.currentWeights[name]

	// 计算总有效权重
	totalWeight := 0
	for i, node := range nodes {
		if node.Enable() {
			weight := node.Weight
			if weight <= 0 {
				weight = 1
			}
			totalWeight += weight
			cw[i] += weight
		}
	}

	if totalWeight == 0 {
		return nil
	}

	// 选择当前权重最大的节点
	bestIdx := -1
	bestWeight := 0
	for i, node := range nodes {
		if node.Enable() && cw[i] > bestWeight {
			bestWeight = cw[i]
			bestIdx = i
		}
	}

	if bestIdx < 0 {
		return nil
	}

	// 被选中的节点当前权重减去总权重
	cw[bestIdx] -= totalWeight
	return nodes[bestIdx]
}
