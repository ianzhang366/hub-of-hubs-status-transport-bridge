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
	MachinepoolPriority             conflationPriority = iota // MinimalComplianceStatus = 5
	KlusterletaddonconfigPriority   conflationPriority = iota // MinimalComplianceStatus = 6
)

// ../dbsyncer/clusterlifecycle_transport_to_db_syncer.go:38,51
func GetPriorityByName(n string) conflationPriority {
	switch n {
	case "clusterdeployments":
		return ClusterDeploymentPriority
	case "machinepools":
		return MachinepoolPriority
	case "klusterletaddonconfigs":
		return KlusterletaddonconfigPriority
	}

	return ManagedClustersPriority
}
