package scheduling

import (
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/scheduling/impl"

	"go.uber.org/zap"
)

type SchedulingHandler interface {
	GetServiceNode(nodes []*models.ServiceNode, name string) *models.ServiceNode
}

func GetHandler(algorithm string, logger *zap.SugaredLogger) SchedulingHandler {
	algo := Algorithm(algorithm)
	switch algo {
	case AlgorithmRandom:
		return impl.NewRandomHandler(logger)
	case AlgorithmRoundRobin:
		return impl.NewRoundHandler(logger)
	// case AlgorithmWeightedRandom:
	// 	return NewWeightedRandomHandler()
	// case AlgorithmIpHash:
	// 	return NewIpHashHandler()
	default:
		return impl.NewRoundHandler(logger)
	}
}

// func Scheduling(nodes []*models.ServiceNode, algorithm Algorithm) *models.ServiceNode {
// 	if len(nodes) == 0 {
// 		return nil
// 	}
// 	if len(nodes) == 1 {
// 		if nodes[0].Enable() {
// 			return nodes[0]
// 		}
// 		return nil
// 	}
// 	switch algorithm {
// 	case AlgorithmRandom:
// 		return nodes[0]
// 	case AlgorithmRoundRobin:
// 		return nodes[0]
// 	// case AlgorithmWeightedRandom:
// 	// 	return nodes[0]
// 	// case AlgorithmIpHash:
// 	// 	return nodes[0]
// 	default:
// 		for _, node := range nodes {
// 			if node.Enable() {
// 				return node
// 			}
// 		}
// 	}
// 	return nil
// }

type Algorithm string

const (
	AlgorithmRandom     Algorithm = "random"
	AlgorithmRoundRobin Algorithm = "robin"
	// AlgorithmWeightedRandom Algorithm = "weight"
	// AlgorithmIpHash         Algorithm = "iphash"
)

func GetAlgorithm(name string) Algorithm {
	switch name {
	case "random":
		return AlgorithmRandom
	case "robin":
		return AlgorithmRoundRobin
	// case "weight":
	// 	return AlgorithmWeightedRandom
	// case "iphash":
	// 	return AlgorithmIpHash
	default:
		return AlgorithmRandom
	}
}
