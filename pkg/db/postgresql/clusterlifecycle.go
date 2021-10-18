package postgresql

import (
	"context"
	"fmt"

	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/db"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/db/postgresql/batch"
)

var _ db.ClusterLifecycleStatusDB = (*PostgreSQL)(nil)

// NewManagedClustersBatchBuilder creates a new instance of ManagedClustersBatchBuilder.
func (p *PostgreSQL) NewClusterLifecycleBatchBuilder(schema string, tableName string,
	leafHubName string) db.ClusterDeploymentBatchBuilder {
	return batch.NewClusterLifecycleBatchBuilder(schema, tableName, leafHubName)
}

// GetClusterDeploymentByLeafHub returns list of clusterdeployments and it's resourceVersion.
func (p *PostgreSQL) GetClusterLifecycleResourceByLeafHub(ctx context.Context, schema string, tableName string,
	leafHubName string) (map[string]string, error) {
	rows, _ := p.conn.Query(ctx, fmt.Sprintf(`SELECT payload->'metadata'->>'name',
		payload->'metadata'->>'resourceVersion' FROM %s.%s WHERE leaf_hub_name=$1`, schema, tableName), leafHubName)

	result := make(map[string]string)

	for rows.Next() {
		clusterName := ""
		resourceVersion := ""

		if err := rows.Scan(&clusterName, &resourceVersion); err != nil {
			return nil, fmt.Errorf("error reading from table %s.%s - %w", schema, tableName, err)
		}

		result[clusterName] = resourceVersion
	}

	return result, nil
}
