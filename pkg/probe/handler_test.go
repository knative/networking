/*
Copyright 2020 The Knative Authors.

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
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"knative.dev/networking/pkg"
	"knative.dev/pkg/network"
	"knative.dev/pkg/network/prober"
	_ "knative.dev/pkg/system/testing"
)

func TestHandlerSuccessful(t *testing.T) {
	body := "Inner Body"
	cases := []struct {
		name    string
		options []interface{}
		want    bool
		expErr  bool
	}{{
		name: "successful probe when both headers are specified",
		options: []interface{}{
			prober.WithHeader(pkg.ProbeHeaderName, headerValue),
			prober.WithHeader(pkg.HashHeaderName, "foo-bar-baz"),
			prober.ExpectsStatusCodes([]int{http.StatusOK}),
		},
		want: true,
	}, {
		name: "forwards to inner handler when probe header is not specified",
		options: []interface{}{
			prober.WithHeader(pkg.HashHeaderName, "foo-bar-baz"),
			prober.ExpectsBody(body),
			// Validates the header is stripped before forwarding to the inner handler
			prober.ExpectsHeader(pkg.HashHeaderName, "false"),
			prober.ExpectsStatusCodes([]int{http.StatusOK}),
		},
		want: true,
	}, {
		name: "forwards to inner handler when probe header is not 'probe'",
		options: []interface{}{
			prober.WithHeader(pkg.ProbeHeaderName, "queue"),
			prober.WithHeader(pkg.HashHeaderName, "foo-bar-baz"),
			prober.ExpectsBody(body),
			prober.ExpectsHeader(pkg.ProbeHeaderName, "true"),
			// Validates the header is stripped before forwarding to the inner handler
			prober.ExpectsHeader(pkg.HashHeaderName, "false"),
			prober.ExpectsStatusCodes([]int{http.StatusOK}),
		},
		want: true,
	}, {
		name: "failed probe when hash header is not present",
		options: []interface{}{
			prober.WithHeader(pkg.ProbeHeaderName, headerValue),
			prober.ExpectsStatusCodes([]int{http.StatusOK}),
		},
		want:   false,
		expErr: true,
	}}

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Header[pkg.ProbeHeaderName]
		w.Header().Set(pkg.ProbeHeaderName, strconv.FormatBool(ok))
		_, ok = r.Header[pkg.HashHeaderName]
		w.Header().Set(pkg.HashHeaderName, strconv.FormatBool(ok))
		w.Write([]byte(body))
	})
	h = NewHandler(h)
	ts := httptest.NewServer(h)
	defer ts.Close()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := prober.Do(context.Background(), network.AutoTransport, ts.URL, c.options...)
			if err != nil && !c.expErr {
				t.Errorf("prober.Do() = %v, no error expected", err)
			}
			if err == nil && c.expErr {
				t.Errorf("prober.Do() = nil, expected an error")
			}
			if got != c.want {
				t.Errorf(" result = %t, want: %t", got, c.want)
			}
		})
	}
}

func BenchmarkHandlerNoHeader(b *testing.B) {
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h = NewHandler(h)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	b.Run("sequential-no-header", func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			h.ServeHTTP(resp, req)
		}
	})

	b.Run("parallel-no-header", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				h.ServeHTTP(resp, req)
			}
		})
	})
}

func BenchmarkHandlerWithHeader(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(pkg.ProbeHeaderName, headerValue)
	req.Header.Set(pkg.HashHeaderName, "ok")
	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h = NewHandler(h)
	b.Run("sequential-probe-header", func(b *testing.B) {
		resp := httptest.NewRecorder()
		for j := 0; j < b.N; j++ {
			h.ServeHTTP(resp, req)
		}
	})

	b.Run("parallel-probe-header", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			// Need to create a separate response writer because of the header mutation at the end of ServeHTTP
			respParallel := httptest.NewRecorder()
			for pb.Next() {
				h.ServeHTTP(respParallel, req)
			}
		})
	})
}
