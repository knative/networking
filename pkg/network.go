/*
Copyright 2018 The Knative Authors

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

package pkg

import (
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	// ProbePath is the name of a path that activator, autoscaler and
	// prober(used by KIngress generally) use for health check.
	ProbePath = "/healthz"

	// ProbeHeaderName is the name of a header that can be added to
	// requests to probe the knative networking layer.  Requests
	// with this header will not be passed to the user container or
	// included in request metrics.
	ProbeHeaderName = "K-Network-Probe"

	// ProxyHeaderName is the name of an internal header that activator
	// uses to mark requests going through it.
	ProxyHeaderName = "K-Proxy-Request"

	// HashHeaderName is the name of an internal header that Ingress controller
	// uses to find out which version of the networking config is deployed.
	HashHeaderName = "K-Network-Hash"

	// HashHeaderValue is the value that must appear in the HashHeaderName
	// header in order for our network hash to be injected.
	HashHeaderValue = "override"

	// OriginalHostHeader is used to avoid Istio host based routing rules
	// in Activator.
	// The header contains the original Host value that can be rewritten
	// at the Queue proxy level back to be a host header.
	OriginalHostHeader = "K-Original-Host"

	// KubeProbeUAPrefix is the user agent prefix of the probe.
	// Since K8s 1.8, prober requests have
	//   User-Agent = "kube-probe/{major-version}.{minor-version}".
	KubeProbeUAPrefix = "kube-probe/"

	// KubeletProbeHeaderName is the name of the header supplied by kubelet
	// probes.  Istio with mTLS rewrites probes, but their probes pass a
	// different user-agent.  So we augment the probes with this header.
	KubeletProbeHeaderName = "K-Kubelet-Probe"

	// UserAgentKey is the constant for header "User-Agent".
	UserAgentKey = "User-Agent"

	// ActivatorUserAgent is the user-agent header value set in probe requests sent
	// from activator.
	ActivatorUserAgent = "Knative-Activator-Probe"

	// QueueProxyUserAgent is the user-agent header value set in probe requests sent
	// from queue-proxy.
	QueueProxyUserAgent = "Knative-Queue-Proxy-Probe"

	// IngressReadinessUserAgent is the user-agent header value
	// set in probe requests for Ingress status.
	IngressReadinessUserAgent = "Knative-Ingress-Probe"

	// AutoscalingUserAgent is the user-agent header value set in probe
	// requests sent by autoscaling implementations.
	AutoscalingUserAgent = "Knative-Autoscaling-Probe"

	// TagHeaderName is the name of the header entry which has a tag name as value.
	// The tag name specifies which route was expected to be chosen by Ingress.
	TagHeaderName = "Knative-Serving-Tag"

	// DefaultRouteHeaderName is the name of the header entry
	// identifying whether a request is routed via the default route or not.
	// It has one of the string value "true" or "false".
	DefaultRouteHeaderName = "Knative-Serving-Default-Route"

	// ProtoAcceptContent is the content type to be used when autoscaler scrapes metrics from the QP
	ProtoAcceptContent = "application/protobuf"

	// FlushInterval controls the time when we flush the connection in the
	// reverse proxies (Activator, QP).
	// As of go1.16, a FlushInterval of 0 (the default) still flushes immediately
	// when Content-Length is -1, which means the default works properly for
	// streaming/websockets, without flushing more often than necessary for
	// non-streaming requests.
	FlushInterval = 0

	// VisibilityLabelKey is the label to indicate visibility of Route
	// and KServices.  It can be an annotation too but since users are
	// already using labels for domain, it probably best to keep this
	// consistent.
	VisibilityLabelKey = "networking.knative.dev/visibility"

	// PassthroughLoadbalancingHeaderName is the name of the header that directs
	// load balancers to not load balance the respective request but to
	// send it to the request's target directly.
	PassthroughLoadbalancingHeaderName = "K-Passthrough-Lb"
)

// IsKubeletProbe returns true if the request is a Kubernetes probe.
func IsKubeletProbe(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), KubeProbeUAPrefix) ||
		r.Header.Get(KubeletProbeHeaderName) != ""
}

// KnativeProbeHeader returns the value for key ProbeHeaderName in request headers.
func KnativeProbeHeader(r *http.Request) string {
	return r.Header.Get(ProbeHeaderName)
}

// KnativeProxyHeader returns the value for key ProxyHeaderName in request headers.
func KnativeProxyHeader(r *http.Request) string {
	return r.Header.Get(ProxyHeaderName)
}

// IsProbe returns true if the request is a Kubernetes probe or a Knative probe,
// i.e. non-empty ProbeHeaderName header.
func IsProbe(r *http.Request) bool {
	return IsKubeletProbe(r) || KnativeProbeHeader(r) != ""
}

// RewriteHostIn removes the `Host` header from the inbound (server) request
// and replaces it with our custom header.
// This is done to avoid Istio Host based routing, see #3870.
// Queue-Proxy will execute the reverse process.
func RewriteHostIn(r *http.Request) {
	h := r.Host
	r.Host = ""
	r.Header.Del("Host")
	// Don't overwrite an existing OriginalHostHeader.
	if r.Header.Get(OriginalHostHeader) == "" {
		r.Header.Set(OriginalHostHeader, h)
	}
}

// RewriteHostOut undoes the `RewriteHostIn` action.
// RewriteHostOut checks if network.OriginalHostHeader was set and if it was,
// then uses that as the r.Host (which takes priority over Request.Header["Host"]).
// If the request did not have the OriginalHostHeader header set, the request is untouched.
func RewriteHostOut(r *http.Request) {
	if ohh := r.Header.Get(OriginalHostHeader); ohh != "" {
		r.Host = ohh
		r.Header.Del("Host")
		r.Header.Del(OriginalHostHeader)
	}
}

// NameForPortNumber finds the name for a given port as defined by a Service.
func NameForPortNumber(svc *corev1.Service, portNumber int32) (string, error) {
	for _, port := range svc.Spec.Ports {
		if port.Port == portNumber {
			return port.Name, nil
		}
	}
	return "", fmt.Errorf("no port with number %d found", portNumber)
}

// PortNumberForName resolves a given name to a portNumber as defined by an EndpointSubset.
func PortNumberForName(sub corev1.EndpointSubset, portName string) (int32, error) {
	for _, subPort := range sub.Ports {
		if subPort.Name == portName {
			return subPort.Port, nil
		}
	}
	return 0, fmt.Errorf("no port for name %q found", portName)
}

// IsPotentialMeshErrorResponse returns whether the HTTP response is compatible
// with having been caused by attempting direct connection when mesh was
// enabled. For example if we get a HTTP 404 status code it's safe to assume
// mesh is not enabled even if a probe was otherwise unsuccessful. This is
// useful to avoid falling back to ClusterIP when we see errors which are
// unrelated to mesh being enabled.
func IsPotentialMeshErrorResponse(resp *http.Response) bool {
	return resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusBadGateway
}
