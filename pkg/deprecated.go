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
	"knative.dev/networking/pkg/httpproxy"
	"knative.dev/networking/pkg/k8s/label"
	"knative.dev/networking/pkg/k8s/port"
	"knative.dev/networking/pkg/mesh"
	"knative.dev/networking/pkg/prober/handler"
)

var (
	// VisibilityLabelKey is the label to indicate visibility of Route
	// and KServices.  It can be an annotation too but since users are
	// already using labels for domain, it probably best to keep this
	// consistent.
	// Deprecated: use label.VisibilityKey
	VisibilityLabelKey = label.VisibilityKey

	// FlushInterval controls the time when we flush the connection in the
	// reverse proxies (Activator, QP).
	// NB: having it equal to 0 is a problem for streaming requests
	// since the data won't be transferred in chunks less than 4kb, if the
	// reverse proxy fails to detect streaming (gRPC, e.g.).
	// Deprecated: use httpproxy.FlushInterval
	FlushInterval = httpproxy.FlushInterval

	// NameForPortNumber finds the name for a given port as defined by a Service.
	// Deprecated: use port.NameForPortNumber
	NameForPortNumber = port.NameForNumber

	// PortNumberForName resolves a given name to a portNumber as defined by an EndpointSubset.
	// Deprecated: use port.NumberForName
	PortNumberForName = port.NumberForName

	// IsPotentialMeshErrorResponse returns whether the HTTP response is compatible
	// with having been caused by attempting direct connection when mesh was
	// enabled. For example if we get a HTTP 404 status code it's safe to assume
	// mesh is not enabled even if a probe was otherwise unsuccessful. This is
	// useful to avoid falling back to ClusterIP when we see errors which are
	// unrelated to mesh being enabled.
	// Deprecated: use mesh.IsPotentialMeshErrorResponse
	IsPotentialMeshErrorResponse = mesh.IsPotentialMeshErrorResponse

	// NewProbeHandler wraps a HTTP handler handling probing requests around the provided HTTP handler
	// Deprecated: use handler.New
	NewProbeHandler = handler.New

	// NewBufferPool creates a new BufferPool. This is only safe to use in the context
	// of a httputil.ReverseProxy, as the buffers returned via Put are not cleaned
	// explicitly.
	// Deprecated: use httpproxy.NewBufferPool
	NewBufferPool = httpproxy.NewBufferPool
)
