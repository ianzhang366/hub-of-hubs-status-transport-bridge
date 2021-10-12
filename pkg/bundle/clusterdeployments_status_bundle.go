package bundle

import cdv1 "github.com/openshift/hive/apis/hive/v1"

// NewClusterdeploymentsStatusBundle creates a new Clusterdeployments status bundle with no data in it.
func NewClusterdeploymentsStatusBundle() *ClusterdeploymentsStatusBundle {
	return &ClusterdeploymentsStatusBundle{}
}

// ClusterdeploymentsStatusBundle abstracts management of ClusterDeployment status bundle.
type ClusterdeploymentsStatusBundle struct {
	Objects     []*cdv1.ClusterDeployment `json:"objects"`
	LeafHubName string                    `json:"leafHubName"`
	Generation  uint64                    `json:"generation"`
}

// GetLeafHubName returns the leaf hub name that sent the bundle.
func (bundle *ClusterdeploymentsStatusBundle) GetLeafHubName() string {
	return bundle.LeafHubName
}

// GetObjects returns the objects in the bundle.
func (bundle *ClusterdeploymentsStatusBundle) GetObjects() []interface{} {
	result := make([]interface{}, len(bundle.Objects))
	for i, obj := range bundle.Objects {
		result[i] = obj
	}

	return result
}

// GetGeneration returns the bundle generation.
func (bundle *ClusterdeploymentsStatusBundle) GetGeneration() uint64 {
	return bundle.Generation
}
