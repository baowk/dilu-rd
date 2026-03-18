package scheduling

import (
	"github.com/baowk/dilu-rd/models"
	"github.com/baowk/dilu-rd/scheduling/impl"
)

type SchedulingHandler interface {
	GetServiceNode(nodes []*models.ServiceNode, name string, clientIP string) *models.ServiceNode
}

func GetHandler(algorithm string) SchedulingHandler {
	algo := Algorithm(algorithm)
	switch algo {
	case AlgorithmRoundRobin:
		return impl.NewRoundRobinHandler()
	case AlgorithmRandom:
		return impl.NewRandomHandler()
	case AlgorithmWeighted:
		return impl.NewWeightedRobinHandler()
	case AlgorithmIpHash:
		return impl.NewIpHashHandler()
	default:
		return impl.NewRoundRobinHandler()
	}
}

type Algorithm string

const (
	AlgorithmRandom     Algorithm = "random"
	AlgorithmRoundRobin Algorithm = "robin"
	AlgorithmWeighted   Algorithm = "weight"
	AlgorithmIpHash     Algorithm = "iphash"
)

func GetAlgorithm(name string) Algorithm {
	switch name {
	case "random":
		return AlgorithmRandom
	case "robin":
		return AlgorithmRoundRobin
	case "weight":
		return AlgorithmWeighted
	case "iphash":
		return AlgorithmIpHash
	default:
		return AlgorithmRandom
	}
}
