module knative.dev/networking

go 1.14

require (
	github.com/google/go-cmp v0.4.0
	go.uber.org/zap v1.14.1
	istio.io/client-go v0.0.0-20200513000250-b1d6e9886b7b
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.0
	k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
	knative.dev/pkg v0.0.0-20200516202602-148827986b00
	knative.dev/sample-controller v0.0.0-20200510050845-bf7c19498b7e
	knative.dev/test-infra v0.0.0-20200515184601-7a28f47cdbcb
)

replace (
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.13+incompatible

	github.com/kubernetes-incubator/custom-metrics-apiserver => github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20190918110929-3d9be26a50eb

	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2

	github.com/tsenart/vegeta => github.com/tsenart/vegeta v1.2.1-0.20190917092155-ab06ddb56e2f

	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
