package batch

import (
	"fmt"
	"strings"

	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/db"
)

const (
	JsonbColumnIndex = 2
	DeleteRowKey     = "payload->'metadata'->>'name'"
)

// NewManagedClustersBatchBuilder creates a new instance of PostgreSQL ManagedClustersBatchBuilder.
func NewClusterDeploymentBatchBuilder(schema string, tableName string, leafHubName string) *ClusterDeploymentBatchBuilder {
	tableSpecialColumns := make(map[int]string)
	tableSpecialColumns[JsonbColumnIndex] = db.Jsonb

	builder := &ClusterDeploymentBatchBuilder{
		baseBatchBuilder: newBaseBatchBuilder(schema, tableName, tableSpecialColumns, leafHubName,
			DeleteRowKey),
	}

	builder.setUpdateStatementFunc(builder.generateUpdateStatement)

	return builder
}

// ManagedClustersBatchBuilder is the PostgreSQL implementation of the ManagedClustersBatchBuilder interface.
type ClusterDeploymentBatchBuilder struct {
	*baseBatchBuilder
}

// Insert adds the given (cluster payload, error string) to the batch to be inserted to the db.
func (builder *ClusterDeploymentBatchBuilder) Insert(payload interface{}, errorString string) {
	builder.insert(builder.leafHubName, payload, errorString)
}

// Update adds the given arguments to the batch to update clusterName with the given payload in db.
func (builder *ClusterDeploymentBatchBuilder) Update(clusterName string, payload interface{}) {
	builder.update(builder.leafHubName, payload, clusterName)
}

// Delete adds delete statement to the batch to delete the given cluster from db.
func (builder *ClusterDeploymentBatchBuilder) Delete(clusterName string) {
	builder.delete(clusterName)
}

// Build builds the batch object.
func (builder *ClusterDeploymentBatchBuilder) Build() interface{} {
	return builder.build()
}

func (builder *ClusterDeploymentBatchBuilder) generateUpdateStatement() string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("UPDATE %s.%s AS old SET payload=new.payload FROM (values ",
		builder.schema, builder.tableName))

	numberOfColumns := len(builder.updateArgs) / builder.updateRowsCount // total num of args divided by num of rows
	stringBuilder.WriteString(builder.generateInsertOrUpdateArgs(builder.updateRowsCount, numberOfColumns,
		builder.tableSpecialColumns))

	stringBuilder.WriteString(") AS new(leaf_hub_name,payload,cluster_name) ")
	stringBuilder.WriteString("WHERE old.leaf_hub_name=new.leaf_hub_name ")
	stringBuilder.WriteString("AND old.payload->'metadata'->>'name'=new.cluster_name")

	return stringBuilder.String()
}
