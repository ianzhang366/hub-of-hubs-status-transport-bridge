package dbsyncer

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/bundle"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/conflator"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/db"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/helpers"
	"github.com/open-cluster-management/hub-of-hubs-status-transport-bridge/pkg/transport"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var errObjectNotClusterdeployment = errors.New("failed to parse object in bundle to a clusterdeployment status")

const (
	ClusterDeployment = "clusterdeployments"
)

// ClusterdeploymentStatusDBSyncer implements clusterdeployment status clusters db sync business logic.
type ClusterlifecycleStatusDBSyncer struct {
	log              logr.Logger
	createBundleFunc func() bundle.Bundle
	component        string
	tableName        string
	msgID            string
}

type option func(*ClusterlifecycleStatusDBSyncer)

func withComponentNameAsTableName(n string) option {
	return func(c *ClusterlifecycleStatusDBSyncer) {
		c.component = n
		c.tableName = n
		c.msgID = n
	}
}

func NewClusterdeploymentStatusDBSyncer(log logr.Logger) DBSyncer {
	return newClusterlifecycleStatusDBSyncer(log, withComponentNameAsTableName(ClusterDeployment))
}

// NewClusterdeploymentStatusDBSyncer creates a new instance of ClusterdeploymentStatusDBSyncer.
func newClusterlifecycleStatusDBSyncer(log logr.Logger, opts ...option) DBSyncer {
	dbSyncer := &ClusterlifecycleStatusDBSyncer{
		log:              log,
		createBundleFunc: func() bundle.Bundle { return bundle.NewClusterlifecycleStatusBundle() },
	}

	log.Info(fmt.Sprintf("initialized %s db syncer", dbSyncer.tableName))

	for _, opt := range opts {
		opt(dbSyncer)
	}

	return dbSyncer
}

// RegisterCreateBundleFunctions registers create bundle functions within the transport instance.
func (syncer *ClusterlifecycleStatusDBSyncer) RegisterCreateBundleFunctions(transportInstance transport.Transport) {
	transportInstance.Register(&transport.BundleRegistration{
		MsgID:            syncer.msgID,
		CreateBundleFunc: syncer.createBundleFunc,
		Predicate:        func() bool { return true }, // by default process all
	})
}

// RegisterBundleHandlerFunctions registers bundle handler functions within the conflation manager.
// handler function need to do "diff" between objects received in the bundle and the objects in db.
// leaf hub sends only the current existing objects, and status transport bridge should understand implicitly which
// objects were deleted.
// therefore, whatever is in the db and cannot be found in the bundle has to be deleted from the db.
// for the objects that appear in both, need to check if something has changed using resourceVersion field comparison
// and if the object was changed, update the db with the current object.
func (syncer *ClusterlifecycleStatusDBSyncer) RegisterBundleHandlerFunctions(conflationManager *conflator.ConflationManager) {
	conflationManager.Register(conflator.NewConflationRegistration(
		conflator.GetPriorityByName(syncer.component),
		helpers.GetBundleType(syncer.createBundleFunc()),
		func(ctx context.Context, bundle bundle.Bundle, dbClient db.StatusTransportBridgeDB) error {
			return syncer.handleBundle(ctx, bundle, dbClient)
		},
	))
}

func (syncer *ClusterlifecycleStatusDBSyncer) handleBundle(ctx context.Context, bundle bundle.Bundle,
	dbClient db.ClusterLifecycleStatusDB) error {
	logBundleHandlingMessage(syncer.log, bundle, startBundleHandlingMessage)
	leafHubName := bundle.GetLeafHubName()

	// ../db/db.go:39
	insFromDB, err := dbClient.GetClusterLifecycleResourceByLeafHub(ctx, db.StatusSchema, syncer.tableName,
		leafHubName)
	if err != nil {
		return fmt.Errorf("failed fetching leaf hub managed clusters from db - %w", err)
	}
	// batch is per leaf hub, therefore no need to specify leafHubName in Insert/Update/Delete
	batchBuilder := dbClient.NewClusterLifecycleBatchBuilder(db.StatusSchema, syncer.tableName, leafHubName)

	for _, object := range bundle.GetObjects() {
		ins, ok := object.(*unstructured.Unstructured)
		if !ok {
			syncer.log.Error(errObjectNotClusterdeployment, "skipping object...")
			continue // do not handle objects other than ManagedCluster
		}

		resourceVersionFromDB, insExistsInDB := insFromDB[ins.GetName()]
		if !insExistsInDB { // clusterdeployments not found in the db table
			batchBuilder.Insert(ins, db.ErrorNone)
			continue
		}

		delete(insFromDB, ins.GetName()) // if we got here, instance exists both in db and in received bundle.

		if ins.GetResourceVersion() == resourceVersionFromDB {
			continue // update cluster in db only if what we got is a different (newer) version of the resource
		}

		batchBuilder.Update(ins.GetName(), ins)
	}
	// delete clusters that in the db but were not sent in the bundle (leaf hub sends only living resources).
	for insName := range insFromDB {
		batchBuilder.Delete(insName)
	}
	// batch contains at most number of statements as the number of managed cluster per LH
	if err := dbClient.SendBatch(ctx, batchBuilder.Build()); err != nil {
		return fmt.Errorf("failed to perform batch - %w", err)
	}

	logBundleHandlingMessage(syncer.log, bundle, finishBundleHandlingMessage)

	return nil
}
