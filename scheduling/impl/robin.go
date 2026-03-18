package impl

import (
	"sync"

	"github.com/baowk/dilu-rd/models"
)

type RoundRobinHandler struct {
	mu  sync.Mutex
	cur map[string]int
}

func NewRoundRobinHandler() *RoundRobinHandler {
	return &RoundRobinHandler{
		cur: make(map[string]int),
	}
}

func (r *RoundRobinHandler) GetServiceNode(nodes []*models.ServiceNode, name string, _ string) *models.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < len(nodes); i++ {
		idx := r.cur[name]
		useIdx := idx % len(nodes)
		r.cur[name] = useIdx + 1
		if nodes[useIdx].Enable() {
			return nodes[useIdx]
		}
	}
	return nil
}
