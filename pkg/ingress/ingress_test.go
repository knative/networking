/*
Copyright 2019 The Knative Authors.

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

package ingress

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func TestGetExpandedHosts(t *testing.T) {
	for _, test := range []struct {
		name  string
		hosts sets.Set[string]
		want  sets.Set[string]
	}{{
		name:  "cluster local service in non-default namespace",
		hosts: sets.New("service.name-space.svc.cluster.local"),
		want: sets.New(
			"service.name-space",
			"service.name-space.svc",
			"service.name-space.svc.cluster.local",
		),
	}, {
		name:  "cluster local service in all-numeric namespace",
		hosts: sets.New("service.1234.svc.cluster.local"),
		want: sets.New(
			"service.1234.svc",
			"service.1234.svc.cluster.local",
		),
	}, {
		name:  "funky namespace",
		hosts: sets.New("service.1-1.svc.cluster.local"),
		want: sets.New(
			"service.1-1",
			"service.1-1.svc",
			"service.1-1.svc.cluster.local",
		),
	}, {
		name: "cluster local service somehow has a very long tld",
		hosts: sets.New(
			"service." + strings.Repeat("s", 64) + ".svc.cluster.local",
		),
		want: sets.New(
			"service."+strings.Repeat("s", 64)+".svc",
			"service."+strings.Repeat("s", 64)+".svc.cluster.local",
		),
	}, {
		name:  "example.com service",
		hosts: sets.New("foo.bar.example.com"),
		want: sets.New(
			"foo.bar.example.com",
		),
	}, {
		name:  "default.example.com service",
		hosts: sets.New("foo.default.example.com"),
		want:  sets.New("foo.default.example.com"),
	}, {
		name: "mix",
		hosts: sets.New(
			"foo.default.example.com",
			"foo.default.svc.cluster.local",
		),
		want: sets.New(
			"foo.default",
			"foo.default.example.com",
			"foo.default.svc",
			"foo.default.svc.cluster.local",
		),
	}} {
		t.Run(test.name, func(t *testing.T) {
			got := ExpandedHosts(test.hosts)
			if !got.Equal(test.want) {
				t.Errorf("ExpandedHosts diff(-want +got):\n%s", cmp.Diff(sets.List(got), sets.List(test.want)))
			}
		})
	}
}

func TestInsertProbe(t *testing.T) {
	tests := []struct {
		name    string
		ingress *v1alpha1.Ingress
		want    string
		wantErr bool
	}{{
		name: "with rules, no append header",
		ingress: &v1alpha1.Ingress{
			Spec: v1alpha1.IngressSpec{
				Rules: []v1alpha1.IngressRule{{
					Hosts: []string{
						"example.com",
					},
					HTTP: &v1alpha1.HTTPIngressRuleValue{
						Paths: []v1alpha1.HTTPIngressPath{{
							Splits: []v1alpha1.IngressBackendSplit{{
								IngressBackend: v1alpha1.IngressBackend{
									ServiceName: "blah",
								},
							}},
						}},
					},
				}},
			},
		},
		want: "a25000a350642c8abef53078b329bd043e18758f6063c1172d53b04e14fcf5c1",
	}, {
		name: "with rules, with append header",
		ingress: &v1alpha1.Ingress{
			Spec: v1alpha1.IngressSpec{
				Rules: []v1alpha1.IngressRule{{
					Hosts: []string{
						"example.com",
					},
					HTTP: &v1alpha1.HTTPIngressRuleValue{
						Paths: []v1alpha1.HTTPIngressPath{{
							Splits: []v1alpha1.IngressBackendSplit{{
								IngressBackend: v1alpha1.IngressBackend{
									ServiceName: "blah",
								},
								AppendHeaders: map[string]string{
									"Foo": "bar",
								},
							}},
						}},
					},
				}},
			},
		},
		want: "6b652c7abed871354affd4a9cb699d33816f24541fac942149b91ad872fe63ca",
	}, {
		name: "rule missing HTTP block",
		ingress: &v1alpha1.Ingress{
			Spec: v1alpha1.IngressSpec{
				Rules: []v1alpha1.IngressRule{{
					Hosts: []string{
						"example.com",
					},
				}},
			},
		},
		wantErr: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ingress := test.ingress.DeepCopy()
			got, err := InsertProbe(test.ingress)
			if test.wantErr == (err == nil) {
				t.Errorf("InsertProbe() err = %v, wantErr = %t", err, test.wantErr)
			}
			if err != nil {
				return
			}
			beforePaths := len(ingress.Spec.Rules[0].HTTP.Paths)
			beforeAppHdr := len(ingress.Spec.Rules[0].HTTP.Paths[0].AppendHeaders)
			beforeMtchHdr := len(ingress.Spec.Rules[0].HTTP.Paths[0].Headers)
			if got != test.want {
				t.Errorf("InsertProbe() = %s, wanted %s", got, test.want)
			}

			afterPaths := len(test.ingress.Spec.Rules[0].HTTP.Paths)
			if beforePaths+beforePaths != afterPaths {
				t.Errorf("InsertProbe() %d paths, wanted %d", afterPaths, beforePaths+beforePaths)
			}

			// Check the matches at the beginning.
			afterAppHdr := len(test.ingress.Spec.Rules[0].HTTP.Paths[0].AppendHeaders)
			if beforeAppHdr+1 != afterAppHdr {
				t.Errorf("InsertProbe() left %d headers, wanted %d", afterAppHdr, beforeAppHdr+1)
			}
			afterMtchHdr := len(test.ingress.Spec.Rules[0].HTTP.Paths[0].Headers)
			if beforeMtchHdr+1 != afterMtchHdr {
				t.Errorf("InsertProbe() left %d header matches, wanted %d", afterMtchHdr, beforeMtchHdr+1)
			}

			// Check the matches at the end
			afterAppHdr = len(test.ingress.Spec.Rules[0].HTTP.Paths[afterPaths-1].AppendHeaders)
			if beforeAppHdr != afterAppHdr {
				t.Errorf("InsertProbe() left %d headers, wanted %d", afterAppHdr, beforeAppHdr)
			}
			afterMtchHdr = len(test.ingress.Spec.Rules[0].HTTP.Paths[afterPaths-1].Headers)
			if beforeMtchHdr != afterMtchHdr {
				t.Errorf("InsertProbe() left %d header matches, wanted %d", afterMtchHdr, beforeMtchHdr)
			}
		})
	}
}

func TestHostsPerVisibility(t *testing.T) {
	tests := []struct {
		name    string
		ingress *v1alpha1.Ingress
		in      map[v1alpha1.IngressVisibility]sets.Set[string]
		want    map[string]sets.Set[string]
	}{{
		name: "external rule",
		ingress: &v1alpha1.Ingress{
			Spec: v1alpha1.IngressSpec{
				Rules: []v1alpha1.IngressRule{{
					Hosts: []string{
						"example.com",
						"foo.bar.svc.cluster.local",
					},
					HTTP: &v1alpha1.HTTPIngressRuleValue{
						Paths: []v1alpha1.HTTPIngressPath{{
							Splits: []v1alpha1.IngressBackendSplit{{
								IngressBackend: v1alpha1.IngressBackend{
									ServiceName: "blah",
								},
								AppendHeaders: map[string]string{
									"Foo": "bar",
								},
							}},
						}},
					},
					Visibility: v1alpha1.IngressVisibilityExternalIP,
				}},
			},
		},
		in: map[v1alpha1.IngressVisibility]sets.Set[string]{
			v1alpha1.IngressVisibilityExternalIP:   sets.New("foo"),
			v1alpha1.IngressVisibilityClusterLocal: sets.New("bar", "baz"),
		},
		want: map[string]sets.Set[string]{
			"foo": sets.New(
				"example.com",
				"foo.bar.svc.cluster.local",
				"foo.bar.svc",
				"foo.bar",
			),
		},
	}, {
		name: "internal rule",
		ingress: &v1alpha1.Ingress{
			Spec: v1alpha1.IngressSpec{
				Rules: []v1alpha1.IngressRule{{
					Hosts: []string{
						"foo.bar.svc.cluster.local",
					},
					HTTP: &v1alpha1.HTTPIngressRuleValue{
						Paths: []v1alpha1.HTTPIngressPath{{
							Splits: []v1alpha1.IngressBackendSplit{{
								IngressBackend: v1alpha1.IngressBackend{
									ServiceName: "blah",
								},
								AppendHeaders: map[string]string{
									"Foo": "bar",
								},
							}},
						}},
					},
					Visibility: v1alpha1.IngressVisibilityClusterLocal,
				}},
			},
		},
		in: map[v1alpha1.IngressVisibility]sets.Set[string]{
			v1alpha1.IngressVisibilityExternalIP:   sets.New("foo"),
			v1alpha1.IngressVisibilityClusterLocal: sets.New("bar", "baz"),
		},
		want: map[string]sets.Set[string]{
			"bar": sets.New(
				"foo.bar.svc.cluster.local",
				"foo.bar.svc",
				"foo.bar",
			),
			"baz": sets.New(
				"foo.bar.svc.cluster.local",
				"foo.bar.svc",
				"foo.bar",
			),
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := HostsPerVisibility(test.ingress, test.in)
			if !cmp.Equal(got, test.want) {
				t.Error("HostsPerVisibility (-want, +got) =", cmp.Diff(test.want, got))
			}
		})
	}
}
