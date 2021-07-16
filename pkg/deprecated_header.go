/*
Copyright 2021 The Knative Authors

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
	"knative.dev/networking/pkg/header"
)

var (
	// ProbePath is the name of a path that activator, autoscaler and
	// prober(used by KIngress generally) use for health check.
	// Deprecated: use header.ProbePath
	ProbePath = header.ProbePath

	// ProbeHeaderName is the name of a header that can be added to
	// requests to probe the knative networking layer.  Requests
	// with this header will not be passed to the user container or
	// included in request metrics.
	// Deprecated: use header.ProbeKey
	ProbeHeaderName = header.ProbeKey

	// ProxyHeaderName is the name of an internal header that activator
	// uses to mark requests going through it.
	// Deprecated: use header.ProxyKey
	ProxyHeaderName = header.ProxyKey

	// HashHeaderName is the name of an internal header that Ingress controller
	// uses to find out which version of the networking config is deployed.
	// Deprecated: use header.HashKey
	HashHeaderName = header.HashKey

	// HashHeaderValue is the value that must appear in the HashHeaderName
	// header in order for our network hash to be injected.
	// Deprecated: use header.HashValue
	HashHeaderValue = header.HashValue

	// OriginalHostHeader is used to avoid Istio host based routing rules
	// in Activator.
	// The header contains the original Host value that can be rewritten
	// at the Queue proxy level back to be a host header.
	// Deprecated: use header.OriginalHostKey
	OriginalHostHeader = header.OriginalHostKey

	// KubeProbeUAPrefix is the user agent prefix of the probe.
	// Since K8s 1.8, prober requests have
	//   User-Agent = "kube-probe/{major-version}.{minor-version}".
	// Deprecated: use header.KubeProbeUAPrefix
	KubeProbeUAPrefix = header.KubeProbeUAPrefix

	// KubeletProbeHeaderName is the name of the header supplied by kubelet
	// probes.  Istio with mTLS rewrites probes, but their probes pass a
	// different user-agent.  So we augment the probes with this header.
	// Deprecated: use header.KubeletProbeKey
	KubeletProbeHeaderName = header.KubeletProbeKey

	// UserAgentKey is the constant for header "User-Agent".
	// Deprecated: use header.UserAgentKey
	UserAgentKey = header.UserAgentKey

	// ActivatorUserAgent is the user-agent header value set in probe requests sent
	// from activator.
	// Deprecated: use header.ActivatorUserAgent
	ActivatorUserAgent = header.ActivatorUserAgent

	// QueueProxyUserAgent is the user-agent header value set in probe requests sent
	// from queue-proxy.
	// Deprecated: use header.QueueProxyUserAgent
	QueueProxyUserAgent = header.QueueProxyUserAgent

	// IngressReadinessUserAgent is the user-agent header value
	// set in probe requests for Ingress status.
	// Deprecated: use header.IngressReadinessUserAgent
	IngressReadinessUserAgent = header.IngressReadinessUserAgent

	// AutoscalingUserAgent is the user-agent header value set in probe
	// requests sent by autoscaling implementations.
	// Deprecated: use header.AutoscalingUserAgent
	AutoscalingUserAgent = header.AutoscalingUserAgent

	// TagHeaderName is the name of the header entry which has a tag name as value.
	// The tag name specifies which route was expected to be chosen by Ingress.
	// Deprecated: use header.RouteTagKey
	TagHeaderName = header.RouteTagKey

	// DefaultRouteHeaderName is the name of the header entry
	// identifying whether a request is routed via the default route or not.
	// It has one of the string value "true" or "false".
	// Deprecated: header.DefaultRouteKey
	DefaultRouteHeaderName = header.DefaultRouteKey

	// ProtoAcceptContent is the content type to be used when autoscaler scrapes metrics from the QP
	// Deprecated: use header.ProtoAcceptContent
	ProtoAcceptContent = header.ProtoAcceptContent

	// PassthroughLoadbalancingHeaderName is the name of the header that directs
	// load balancers to not load balance the respective request but to
	// send it to the request's target directly.
	// Deprecated: use header.PassthroughLoadbalancingKey
	PassthroughLoadbalancingHeaderName = header.PassthroughLoadbalancingKey

	// IsKubeletProbe returns true if the request is a Kubernetes probe.
	// Deprecated: use header.IsKubeletProbe
	IsKubeletProbe = header.IsKubeletProbe

	// KnativeProbeHeader returns the value for key ProbeHeaderName in request headers.
	// Deprecated: use header.GetKnativeProbeValue
	KnativeProbeHeader = header.GetKnativeProbeValue

	// KnativeProxyHeader returns the value for key ProxyHeaderName in request headers.
	// Deprecated: use header.GetKnativeProxyValue
	KnativeProxyHeader = header.GetKnativeProxyValue

	// IsProbe returns true if the request is a Kubernetes probe or a Knative probe,
	// i.e. non-empty ProbeHeaderName header.
	// Deprecated: header.IsProbe
	IsProbe = header.IsProbe

	// RewriteHostIn removes the `Host` header from the inbound (server) request
	// and replaces it with our custom header.
	// This is done to avoid Istio Host based routing, see #3870.
	// Queue-Proxy will execute the reverse process.
	// Deprecated: use header.RewriteHostIn
	RewriteHostIn = header.RewriteHostIn

	// RewriteHostOut undoes the `RewriteHostIn` action.
	// RewriteHostOut checks if network.OriginalHostHeader was set and if it was,
	// then uses that as the r.Host (which takes priority over Request.Header["Host"]).
	// If the request did not have the OriginalHostHeader header set, the request is untouched.
	// Deprecated: use header.RewriteHostOut
	RewriteHostOut = header.RewriteHostOut

	// ProbeHeaderValue is the value used in 'K-Network-Probe'
	// Deprecated: use header.ProbeValue
	ProbeHeaderValue = header.ProbeValue
)
