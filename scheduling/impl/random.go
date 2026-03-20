package impl

import (
	"math/rand"
	"sync"
	"time"

	"github.com/baowk/dilu-rd/models"
)

type RandomHandler struct {
	mu sync.Mutex
	r  *rand.Rand
}

func NewRandomHandler() *RandomHandler {
	return &RandomHandler{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (rh *RandomHandler) GetServiceNode(nodes []*models.ServiceNode, name string, _ string) *models.ServiceNode {
	if len(nodes) == 0 {
		return nil
	}

	n := len(nodes)
	rh.mu.Lock()
	start := rh.r.Intn(n)
	rh.mu.Unlock()

	for i := 0; i < n; i++ {
		node := nodes[(start+i)%n]
		if node.Enable() {
			return node
		}
	}
	return nil
}
