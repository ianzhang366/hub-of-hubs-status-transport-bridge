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
	cdv1 "github.com/openshift/hive/apis/hive/v1"
)

const (
	ClusterDeploymentMsgKey = "clusterdeployments"
)

var errObjectNotClusterdeployment = errors.New("failed to parse object in bundle to a clusterdeployment status")

// NewClusterdeploymentStatusDBSyncer creates a new instance of ClusterdeploymentStatusDBSyncer.
func NewClusterdeploymentStatusDBSyncer(log logr.Logger) DBSyncer {
	dbSyncer := &ClusterdeploymentStatusDBSyncer{
		log:              log,
		createBundleFunc: func() bundle.Bundle { return bundle.NewClusterdeploymentsStatusBundle() },
	}

	log.Info(fmt.Sprintf("initialized %s db syncer", db.ClusterDeploymentTable))

	return dbSyncer
}

// ClusterdeploymentStatusDBSyncer implements clusterdeployment status clusters db sync business logic.
type ClusterdeploymentStatusDBSyncer struct {
	log              logr.Logger
	createBundleFunc func() bundle.Bundle
}

// RegisterCreateBundleFunctions registers create bundle functions within the transport instance.
func (syncer *ClusterdeploymentStatusDBSyncer) RegisterCreateBundleFunctions(transportInstance transport.Transport) {
	transportInstance.Register(&transport.BundleRegistration{
		MsgID:            db.ClusterDeploymentTable,
		CreateBundleFunc: syncer.createBundleFunc,
		Predicate:        func() bool { return true }, // always get managed clusters bundles
	})
}

// RegisterBundleHandlerFunctions registers bundle handler functions within the conflation manager.
// handler function need to do "diff" between objects received in the bundle and the objects in db.
// leaf hub sends only the current existing objects, and status transport bridge should understand implicitly which
// objects were deleted.
// therefore, whatever is in the db and cannot be found in the bundle has to be deleted from the db.
// for the objects that appear in both, need to check if something has changed using resourceVersion field comparison
// and if the object was changed, update the db with the current object.
func (syncer *ClusterdeploymentStatusDBSyncer) RegisterBundleHandlerFunctions(conflationManager *conflator.ConflationManager) {
	conflationManager.Register(conflator.NewConflationRegistration(
		conflator.ClusterDeploymentPriority,
		helpers.GetBundleType(syncer.createBundleFunc()),
		func(ctx context.Context, bundle bundle.Bundle, dbClient db.StatusTransportBridgeDB) error {
			return syncer.handleBundle(ctx, bundle, dbClient)
		},
	))
}

func (syncer *ClusterdeploymentStatusDBSyncer) handleBundle(ctx context.Context, bundle bundle.Bundle,
	dbClient db.ClusterDeploymentStatusDB) error {
	logBundleHandlingMessage(syncer.log, bundle, startBundleHandlingMessage)
	leafHubName := bundle.GetLeafHubName()

	cdFromDB, err := dbClient.GetClusterDeploymentByLeafHub(ctx, db.StatusSchema, db.ClusterDeploymentTable,
		leafHubName)
	if err != nil {
		return fmt.Errorf("failed fetching leaf hub managed clusters from db - %w", err)
	}
	// batch is per leaf hub, therefore no need to specify leafHubName in Insert/Update/Delete
	batchBuilder := dbClient.NewClusterDeploymentBatchBuilder(db.StatusSchema, db.ClusterDeploymentTable, leafHubName)

	for _, object := range bundle.GetObjects() {
		cd, ok := object.(*cdv1.ClusterDeployment)
		if !ok {
			syncer.log.Error(errObjectNotClusterdeployment, "skipping object...")
			continue // do not handle objects other than ManagedCluster
		}

		resourceVersionFromDB, cdExistsInDB := cdFromDB[cd.GetName()]
		if !cdExistsInDB { // clusterdeployments not found in the db table
			batchBuilder.Insert(cd, db.ErrorNone)
			continue
		}

		delete(cdFromDB, cd.GetName()) // if we got here, clusterdeployments exists both in db and in received bundle.

		if cd.GetResourceVersion() == resourceVersionFromDB {
			continue // update cluster in db only if what we got is a different (newer) version of the resource
		}

		batchBuilder.Update(cd.GetName(), cd)
	}
	// delete clusters that in the db but were not sent in the bundle (leaf hub sends only living resources).
	for cdName := range cdFromDB {
		batchBuilder.Delete(cdName)
	}
	// batch contains at most number of statements as the number of managed cluster per LH
	if err := dbClient.SendBatch(ctx, batchBuilder.Build()); err != nil {
		return fmt.Errorf("failed to perform batch - %w", err)
	}

	logBundleHandlingMessage(syncer.log, bundle, finishBundleHandlingMessage)

	return nil
}
