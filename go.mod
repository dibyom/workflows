module github.com/tektoncd/experimental/workflows

go 1.16

require (
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20210729182058-ea1f5c7c37f1
	github.com/google/go-cmp v0.5.6
	github.com/spf13/cobra v1.2.1
	github.com/tektoncd/pipeline v0.32.1
	github.com/tektoncd/plumbing v0.0.0-20211012143332-c7cc43d9bc0c
	github.com/tektoncd/triggers v0.18.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v0.22.5
	k8s.io/code-generator v0.22.5
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
	knative.dev/pkg v0.0.0-20220104185830-52e42b760b54
	sigs.k8s.io/yaml v1.3.0
)

// Knative deps (release-0.20)
replace (
	contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
)
