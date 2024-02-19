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
	case AlgorithmRoundRobin:
		return impl.NewRoundRobinHandler(logger)
	case AlgorithmRandom:
		return impl.NewRandomHandler(logger)
	// case AlgorithmWeightedRandom:
	// 	return NewWeightedRandomHandler()
	// case AlgorithmIpHash:
	// 	return NewIpHashHandler()
	default:
		return impl.NewRoundRobinHandler(logger)
	}
}

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
