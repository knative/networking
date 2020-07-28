/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"net/http"
	"strings"
)

const (
	// KubeProbeUAPrefix is the user agent prefix of the probe.
	// Since K8s 1.8, prober requests have
	//   User-Agent = "kube-probe/{major-version}.{minor-version}".
	KubeProbeUAPrefix = "kube-probe/"

	// KubeletProbeHeaderName is the name of the header supplied by kubelet
	// probes.  Istio with mTLS rewrites probes, but their probes pass a
	// different user-agent.  So we augment the probes with this header.
	KubeletProbeHeaderName = "K-Kubelet-Probe"
)

// IsKubeletProbe returns true if the request is a Kubernetes probe.
func IsKubeletProbe(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), KubeProbeUAPrefix) ||
		r.Header.Get(KubeletProbeHeaderName) != ""
}
