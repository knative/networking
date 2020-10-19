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

package environment

import (
	"flag"
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Cluster struct {
	Name       string // K8s cluster (defaults to cluster in kubeconfig)
	KubeConfig string // Path to kubeconfig (defaults to ./kube/config)
	DomainName string // ie. cluster.local
}

func (s *Cluster) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.Name, "env.cluster.name", "",
		"Provide the cluster to test against. Defaults to the current cluster in kubeconfig.")

	// Allow for --env.cluster.kubeconfig to override the KUBECONFIG env var
	fs.StringVar(&s.KubeConfig, "env.cluster.kubeconfig", defaultKubeConfig(),
		"Provide the path to the `kubeconfig` file you'd like to use for these tests. The `current-context` will be used.")

	// Allows the test to use a different cluster domain - ie. if it's not cluster.local
	fs.StringVar(&s.DomainName, "env.cluster.domain", "cluster.local",
		"Provide the path to the `kubeconfig` file you'd like to use for these tests. The `current-context` will be used.")
}

func defaultKubeConfig() string {
	// Use KUBECONFIG if available
	kubeconfig := os.Getenv("KUBECONFIG")

	// If KUBECONFIG env var isn't set then look for $HOME/.kube/config
	if kubeconfig == "" {
		if usr, err := user.Current(); err == nil {
			kubeconfig = path.Join(usr.HomeDir, ".kube/config")
		}
	}
	return kubeconfig
}

func (c *Cluster) ClientConfig() *rest.Config {
	overrides := &clientcmd.ConfigOverrides{}
	overrides.Context.Cluster = c.Name

	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: c.KubeConfig}

	conf, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides).ClientConfig()
	if err != nil {
		panic(err)
	}
	return conf
}
