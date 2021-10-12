module github.com/open-cluster-management/hub-of-hubs-status-transport-bridge

go 1.16

require (
	github.com/deckarep/golang-set v1.7.1
	github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/jackc/pgx/v4 v4.11.0
	github.com/open-cluster-management/api v0.0.0-20210527013639-a6845f2ebcb1
	github.com/open-cluster-management/hub-of-hubs-data-types v0.1.0
	github.com/open-horizon/edge-sync-service-client v0.0.0-20190711093406-dc3a19905da2
	github.com/open-horizon/edge-utilities v0.0.0-20190711093331-0908b45a7152 // indirect
	github.com/openshift/hive/apis v0.0.0-20211012143010-16ef5a35537d
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/spf13/pflag v1.0.5
	k8s.io/apimachinery v0.20.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.2
)

replace k8s.io/client-go => k8s.io/client-go v0.20.5
