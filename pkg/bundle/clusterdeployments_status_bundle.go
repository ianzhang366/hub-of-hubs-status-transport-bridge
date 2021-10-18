package bundle

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// NewClusterlifecycleStatusBundle creates a new Clusterlifecycle resource bundle template with no data in it.
func NewClusterlifecycleStatusBundle() *ClusterlifecycleStatusBundle {
	return &ClusterlifecycleStatusBundle{}
}

// ClusterdeploymentsStatusBundle abstracts management of ClusterDeployment status bundle.
type ClusterlifecycleStatusBundle struct {
	Objects     []*unstructured.Unstructured `json:"objects"`
	LeafHubName string                       `json:"leafHubName"`
	Generation  uint64                       `json:"generation"`
}

// GetLeafHubName returns the leaf hub name that sent the bundle.
func (bundle *ClusterlifecycleStatusBundle) GetLeafHubName() string {
	return bundle.LeafHubName
}

// GetObjects returns the objects in the bundle.
func (bundle *ClusterlifecycleStatusBundle) GetObjects() []interface{} {
	result := make([]interface{}, len(bundle.Objects))
	for i, obj := range bundle.Objects {
		result[i] = obj
	}

	return result
}

// GetGeneration returns the bundle generation.
func (bundle *ClusterlifecycleStatusBundle) GetGeneration() uint64 {
	return bundle.Generation
}
