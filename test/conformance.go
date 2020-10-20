/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"flag"
	"os"
	"strings"
	"testing"

	network "knative.dev/networking/pkg"
	"knative.dev/pkg/system"
	"knative.dev/reconciler-test/pkg/test"
	"knative.dev/reconciler-test/pkg/test/environment"

	// Mysteriously required to support GCP auth (required by k8s libs). Apparently just importing it is enough. @_@ side effects @_@. https://github.com/kubernetes/client-go/issues/242
	"k8s.io/apimachinery/pkg/util/sets"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// Allow E2E to run against a cluster using OpenID.
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Constants for test images located in test/test_images.
const (
	// Test image names
	Autoscale           = "autoscale"
	Failing             = "failing"
	HelloVolume         = "hellovolume"
	HelloWorld          = "helloworld"
	HTTPProxy           = "httpproxy"
	InvalidHelloWorld   = "invalidhelloworld" // Not a real image
	PizzaPlanet1        = "pizzaplanetv1"
	PizzaPlanet2        = "pizzaplanetv2"
	Protocols           = "protocols"
	Runtime             = "runtime"
	SingleThreadedImage = "singlethreaded"
	Timeout             = "timeout"
	WorkingDir          = "workingdir"

	// Constants for test image output.
	PizzaPlanetText1 = "What a spaceport!"
	PizzaPlanetText2 = "Re-energize yourself with a slice of pepperoni!"
	HelloWorldText   = "Hello World! How about some tasty noodles?"

	ConcurrentRequests = 200
	// We expect to see 100% of requests succeed for traffic sent directly to revisions.
	// This might be a bad assumption.
	MinDirectPercentage = 1
	// We expect to see at least 25% of either response since we're routing 50/50.
	// The CDF of the binomial distribution tells us this will flake roughly
	// 1 time out of 10^12 (roughly the number of galaxies in the observable universe).
	MinSplitPercentage = 0.25
)

var Init = test.Init

type T struct {
	test.T

	Cluster environment.Cluster
	Images  environment.Images
	Ingress environment.Ingress
	Spoof   environment.Spoof

	Clients *Clients

	TestNamespace   string // The test namespace were test resources are created
	SystemNamespace string // The system namespace where the control plane is running

	ResolvableDomain bool        // Resolve Route controller's `domainSuffix`
	HTTPS            bool        // Indicates where the test service will be created with https
	IngressClass     string      // Indicates the class of Ingress provider to test.
	CertificateClass string      // Indicates the class of Certificate provider to test.
	Buckets          int         // The number of reconciler buckets configured.
	Replicas         int         // The number of controlplane replicas being run.
	SkipTests        sets.String // Indicates the test names we want to skip in alpha or beta features.
}

func (t *T) AddFlags(fs *flag.FlagSet) {
	t.T.AddFlags(fs)

	t.Cluster.AddFlags(fs)
	t.Images.AddFlags(fs)
	t.Ingress.AddFlags(fs)
	t.Spoof.AddFlags(fs)

	fs.StringVar(&t.TestNamespace, "env.test-namespace", ServingNamespace,
		"Provide the namespace where test resources will be created")

	systemNamespace := os.Getenv(system.NamespaceEnvKey)
	if systemNamespace == "" {
		systemNamespace = "knative-serving"
	}

	fs.StringVar(&t.SystemNamespace, "env.system-namespace", systemNamespace,
		"Provide the namespace where test resources will be created")

	flag.BoolVar(&t.ResolvableDomain,
		"resolvabledomain",
		false,
		"Set this flag to true if you have configured the `domainSuffix` on your Route controller to a domain that will resolve to your test cluster.")

	flag.BoolVar(&t.HTTPS,
		"https",
		false,
		"Set this flag to true to run all tests with https.")

	flag.StringVar(&t.IngressClass,
		"ingressClass",
		network.IstioIngressClassName,
		"Set this flag to the ingress class to test against.")

	flag.StringVar(&t.CertificateClass,
		"certificateClass",
		network.CertManagerCertificateClassName,
		"Set this flag to the certificate class to test against.")

	flag.IntVar(&t.Buckets,
		"buckets",
		1,
		"Set this flag to the number of reconciler buckets configured.")

	flag.IntVar(&t.Replicas,
		"replicas",
		1,
		"Set this flag to the number of controlplane replicas being run.")

	flag.Var(commaVar{&t.SkipTests},
		"skip-tests",
		"Set this flag to the tests you want to skip in alpha or beta features. Accepts a comma separated list.")
}

func (t *T) Setup(tt *testing.T) {
	cfg := t.Cluster.ClientConfig()
	cfg.QPS = 100
	cfg.Burst = 200

	// TODO - setup logstream
	var err error
	t.Clients, err = NewClientsFromConfig(cfg, t.TestNamespace)

	if err != nil {
		t.Fatal("Failed to create clients", err)
	}
}

func (t *T) Alpha(name string, f interface{}) bool {
	if t.SkipTests.Has(name) {
		f = func(t *test.T) {
			t.Skip("test explicitly skipped via flag")
		}
	}
	return t.T.Alpha(name, f)
}

func (t *T) Beta(name string, f interface{}) bool {
	if t.SkipTests.Has(name) {
		f = func(t *test.T) {
			t.Skip("test explicitly skipped via flag")
		}
	}
	return t.T.Beta(name, f)
}

type commaVar struct {
	set *sets.String
}

func (c commaVar) Set(value string) error {
	vals := strings.Split(value, ",")
	*c.set = sets.NewString(vals...)
	return nil
}

func (c commaVar) String() string {
	return strings.Join(c.set.List(), ", ")
}
