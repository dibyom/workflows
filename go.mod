module github.com/tektoncd/experimental/workflows

go 1.16

require (
	github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher v0.0.0-20210729182058-ea1f5c7c37f1
	github.com/google/go-cmp v0.5.7
	github.com/spf13/cobra v1.3.0
	github.com/tektoncd/pipeline v0.35.1
	github.com/tektoncd/plumbing v0.0.0-20220329085922-d765a5cba75f
	github.com/tektoncd/triggers v0.20.1
	go.uber.org/zap v1.19.1
	k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver v0.23.4
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	k8s.io/code-generator v0.23.5
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf
	knative.dev/pkg v0.0.0-20220329144915-0a1ec2e0d46c
	sigs.k8s.io/yaml v1.3.0
)

// Knative deps (release-0.20)
replace (
	contrib.go.opencensus.io/exporter/stackdriver => contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
)
