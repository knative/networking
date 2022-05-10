/*
Copyright 2018 The Knative Authors.

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

package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	_ "knative.dev/pkg/system/testing"
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
	req.Header.Set(KubeletProbeHeaderName, "no matter")
	if !IsKubeletProbe(req) {
		t.Error("kubelet probe but not counted as such")
	}
}

func TestKnativeProbeHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if h := KnativeProbeHeader(req); h != "" {
		t.Errorf("KnativeProbeHeader(req)=%v, want empty string", h)
	}
	const want = "activator"
	req.Header.Set(ProbeHeaderName, want)
	if h := KnativeProbeHeader(req); h != want {
		t.Errorf("KnativeProbeHeader(req)=%v, want %v", h, want)
	}
	req.Header.Set(ProbeHeaderName, "")
	if h := KnativeProbeHeader(req); h != "" {
		t.Errorf("KnativeProbeHeader(req)=%v, want empty string", h)
	}
}

func TestKnativeProxyHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal("Error building request:", err)
	}
	if h := KnativeProxyHeader(req); h != "" {
		t.Errorf("KnativeProxyHeader(req)=%v, want empty string", h)
	}
	const want = "activator"
	req.Header.Set(ProxyHeaderName, want)
	if h := KnativeProxyHeader(req); h != want {
		t.Errorf("KnativeProxyHeader(req)=%v, want %v", h, want)
	}
	req.Header.Set(ProxyHeaderName, "")
	if h := KnativeProxyHeader(req); h != "" {
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
	req.Header.Set(ProbeHeaderName, "activator")
	if !IsProbe(req) {
		t.Error("Knative probe but not counted as such")
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

	if got, want := r.Header.Get(OriginalHostHeader), "love.is"; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostHeader, got, want)
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

	if got, want := r.Header.Get(OriginalHostHeader), "love.is"; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostHeader, got, want)
	}

	RewriteHostOut(r)
	if got, want := r.Host, "love.is"; got != want {
		t.Errorf("r.Host = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get("Host"), ""; got != want {
		t.Errorf("r.Header['Host'] = %q, want: %q", got, want)
	}

	if got, want := r.Header.Get(OriginalHostHeader), ""; got != want {
		t.Errorf("r.Header[%s] = %q, want: %q", OriginalHostHeader, got, want)
	}
}

func TestNameForPortNumber(t *testing.T) {
	for _, tc := range []struct {
		name       string
		svc        *corev1.Service
		portNumber int32
		portName   string
		err        error
	}{{
		name: "HTTP to 80",
		svc: &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 80,
					Name: "http",
				}, {
					Port: 443,
					Name: "https",
				}},
			},
		},
		portName:   "http",
		portNumber: 80,
	}, {
		name: "no port",
		svc: &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 443,
					Name: "https",
				}},
			},
		},
		portNumber: 80,
		err:        errors.New("no port with number 80 found"),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			portName, err := NameForPortNumber(tc.svc, tc.portNumber)
			if !reflect.DeepEqual(err, tc.err) { // cmp Doesn't work well here due to private fields.
				t.Errorf("Err = %v, want: %v", err, tc.err)
			}
			if tc.err == nil && portName != tc.portName {
				t.Errorf("PortName = %s, want: %s", portName, tc.portName)
			}
		})
	}
}

func TestPortNumberForName(t *testing.T) {
	for _, tc := range []struct {
		name       string
		subset     corev1.EndpointSubset
		portNumber int32
		portName   string
		err        error
	}{{
		name: "HTTP to 80",
		subset: corev1.EndpointSubset{
			Ports: []corev1.EndpointPort{{
				Port: 8080,
				Name: "http",
			}, {
				Port: 8443,
				Name: "https",
			}},
		},
		portName:   "http",
		portNumber: 8080,
	}, {
		name: "no port",
		subset: corev1.EndpointSubset{
			Ports: []corev1.EndpointPort{{
				Port: 8443,
				Name: "https",
			}},
		},
		portName: "http",
		err:      errors.New(`no port for name "http" found`),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			portNumber, err := PortNumberForName(tc.subset, tc.portName)
			if !reflect.DeepEqual(err, tc.err) { // cmp Doesn't work well here due to private fields.
				t.Errorf("Err = %v, want: %v", err, tc.err)
			}
			if tc.err == nil && portNumber != tc.portNumber {
				t.Errorf("PortNumber = %d, want: %d", portNumber, tc.portNumber)
			}
		})
	}
}

func TestIsPotentialMeshErrorResponse(t *testing.T) {
	for _, test := range []struct {
		statusCode int
		expect     bool
	}{{
		statusCode: 404,
		expect:     false,
	}, {
		statusCode: 200,
		expect:     false,
	}, {
		statusCode: 501,
		expect:     false,
	}, {
		statusCode: 502,
		expect:     true,
	}, {
		statusCode: 503,
		expect:     true,
	}} {
		t.Run(fmt.Sprintf("statusCode=%d", test.statusCode), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: test.statusCode,
			}
			if got := IsPotentialMeshErrorResponse(resp); got != test.expect {
				t.Errorf("IsPotentialMeshErrorResponse({StatusCode: %d}) = %v, expected %v", resp.StatusCode, got, test.expect)
			}
		})
	}
}
