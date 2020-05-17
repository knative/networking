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

package probe

import (
	"fmt"
	"net/http"
)

const (
	// HeaderName is the name of a header that can be added to
	// requests to probe the knative networking layer.  Requests
	// with this header will not be passed to the user container or
	// included in request metrics.
	HeaderName = "K-Network-"

	// ProxyHeaderName is the name of an internal header that activator
	// uses to mark requests going through it.
	ProxyHeaderName = "K-Proxy-Request"

	// HashHeaderName is the name of an internal header that Ingress controller
	// uses to find out which version of the networking config is deployed.
	HashHeaderName = "K-Network-Hash"
)

// HeaderValue is the value used in 'K-Network-'
var HeaderValue = "probe"

type handler struct {
	next http.Handler
}

// NewHandler wraps a HTTP handler handling probing requests around the provided HTTP handler
func NewHandler(next http.Handler) http.Handler {
	return &handler{next: next}
}

// ServeHTTP handles probing requests
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ph := r.Header.Get(HeaderName); ph != HeaderValue {
		r.Header.Del(HashHeaderName)
		h.next.ServeHTTP(w, r)
		return
	}

	hh := r.Header.Get(HashHeaderName)
	if hh == "" {
		http.Error(w, fmt.Sprintf("a probe request must contain a non-empty %q header", HashHeaderName), http.StatusBadRequest)
		return
	}

	w.Header().Set(HashHeaderName, hh)
	w.WriteHeader(200)
}
