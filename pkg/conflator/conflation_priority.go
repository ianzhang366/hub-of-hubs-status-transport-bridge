package conflator

type conflationPriority uint

// iota is the golang way of implementing ENUM

// priority list of conflation unit.
const (
	ManagedClustersPriority         conflationPriority = iota // ManagedClusters = 0
	ClustersPerPolicyPriority       conflationPriority = iota // ClustersPerPolicy = 1
	ComplianceStatusPriority        conflationPriority = iota // ComplianceStatus = 2
	MinimalComplianceStatusPriority conflationPriority = iota // MinimalComplianceStatus = 3
	ClusterDeploymentPriority       conflationPriority = iota // MinimalComplianceStatus = 4
)

func GetPriorityByName(n string) conflationPriority {
	switch n {
	case "clusterdeployments":
		return ClusterDeploymentPriority
	}

	return ManagedClustersPriority
}
