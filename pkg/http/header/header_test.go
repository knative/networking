/*
Copyright 2022 The Knative Authors

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

package header

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsKubeletProbe(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if IsKubeletProbe(req) {
		t.Error("Not a kubelet probe but counted as such")
	}
	req.Header.Set("User-Agent", KubeProbeUAPrefix+"1.14")
	if !IsKubeletProbe(req) {
		t.Error("kubelet probe but not counted as such")
	}
	req.Header.Del("User-Agent")
	if IsKubeletProbe(req) {
		t.Error("Not a kubelet probe but counted as such")
	}
	req.Header.Set(KubeletProbeKey, "no matter")
	if !IsKubeletProbe(req) {
		t.Error("kubelet probe but not counted as such")
	}
}

func TestKnativeProbeHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if h := GetKnativeProbeValue(req); h != "" {
		t.Errorf("KnativeProbeHeader(req)=%v, want empty string", h)
	}
	const want = "activator"
	req.Header.Set(ProbeKey, want)
	if h := GetKnativeProbeValue(req); h != want {
		t.Errorf("KnativeProbeHeader(req)=%v, want %v", h, want)
	}
	req.Header.Set(ProbeKey, "")
	if h := GetKnativeProbeValue(req); h != "" {
		t.Errorf("KnativeProbeHeader(req)=%v, want empty string", h)
	}
}

func TestKnativeProxyHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if h := GetKnativeProxyValue(req); h != "" {
		t.Errorf("KnativeProxyHeader(req)=%v, want empty string", h)
	}
	const want = "activator"
	req.Header.Set(ProxyKey, want)
	if h := GetKnativeProxyValue(req); h != want {
		t.Errorf("KnativeProxyHeader(req)=%v, want %v", h, want)
	}
	req.Header.Set(ProxyKey, "")
	if h := GetKnativeProxyValue(req); h != "" {
		t.Errorf("KnativeProxyHeader(req)=%v, want empty string", h)
	}
}

func TestIsProbe(t *testing.T) {
	// Not a probe
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if IsProbe(req) {
		t.Error("Not a probe but counted as such")
	}
	// Kubelet probe
	req.Header.Set("User-Agent", KubeProbeUAPrefix+"1.14")
	if !IsProbe(req) {
		t.Error("Kubelet probe but not counted as such")
	}
	// Knative probe
	req.Header.Del("User-Agent")
	req.Header.Set(ProbeKey, "activator")
	if !IsProbe(req) {
		t.Error("Knative probe but not counted as such")
	}
}

func TestIsMetricsIgnored(t *testing.T) {
	// Not a metrics ignored
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if IsMetricsIgnored(req) {
		t.Error("Not a metrics ignored but counted as such")
	}
	req.Header.Set(IgnoreMetricsKey, "true")
	if !IsMetricsIgnored(req) {
		t.Error("Metrics ignored but not counted as such")
	}
}

func TestRewriteHost(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://love.is/not-hate", nil)
	r.Header.Set("Host", "love.is")

	RewriteHostIn(r)

	if got, want := r.Host, ""; got != want {
		t.Errorf("r.Host = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get("Host"), ""; got != want {
		t.Errorf("r.Header['Host'] = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get(OriginalHostKey), "love.is"; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostKey, got, want)
	}

	// Do it again, but make sure that the ORIGINAL domain is still preserved.
	r.Header.Set("Host", "hate.is")
	RewriteHostIn(r)

	if got, want := r.Host, ""; got != want {
		t.Errorf("r.Host = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get("Host"), ""; got != want {
		t.Errorf("r.Header['Host'] = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get(OriginalHostKey), "love.is"; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostKey, got, want)
	}

	RewriteHostOut(r)
	if got, want := r.Host, "love.is"; got != want {
		t.Errorf("r.Host = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get("Host"), ""; got != want {
		t.Errorf("r.Header['Host'] = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get(OriginalHostKey), ""; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostKey, got, want)
	}
}
