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
		hosts sets.String
		want  sets.String
	}{{
		name:  "cluster local service in non-default namespace",
		hosts: sets.NewString("service.name-space.svc.cluster.local"),
		want: sets.NewString(
			"service.name-space",
			"service.name-space.svc",
			"service.name-space.svc.cluster.local",
		),
	}, {
		name:  "cluster local service in all-numeric namespace",
		hosts: sets.NewString("service.1234.svc.cluster.local"),
		want: sets.NewString(
			"service.1234.svc",
			"service.1234.svc.cluster.local",
		),
	}, {
		name:  "funky namespace",
		hosts: sets.NewString("service.1-1.svc.cluster.local"),
		want: sets.NewString(
			"service.1-1",
			"service.1-1.svc",
			"service.1-1.svc.cluster.local",
		),
	}, {
		name: "cluster local service somehow has a very long tld",
		hosts: sets.NewString(
			"service." + strings.Repeat("s", 64) + ".svc.cluster.local",
		),
		want: sets.NewString(
			"service."+strings.Repeat("s", 64)+".svc",
			"service."+strings.Repeat("s", 64)+".svc.cluster.local",
		),
	}, {
		name:  "example.com service",
		hosts: sets.NewString("foo.bar.example.com"),
		want: sets.NewString(
			"foo.bar.example.com",
		),
	}, {
		name:  "default.example.com service",
		hosts: sets.NewString("foo.default.example.com"),
		want:  sets.NewString("foo.default.example.com"),
	}, {
		name: "mix",
		hosts: sets.NewString(
			"foo.default.example.com",
			"foo.default.svc.cluster.local",
		),
		want: sets.NewString(
			"foo.default",
			"foo.default.example.com",
			"foo.default.svc",
			"foo.default.svc.cluster.local",
		),
	}} {
		t.Run(test.name, func(t *testing.T) {
			got := ExpandedHosts(test.hosts)
			if !got.Equal(test.want) {
				t.Errorf("ExpandedHosts diff(-want +got):\n%s", cmp.Diff(got.List(), test.want.List()))
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
		want: "f27ea4155f65534151b83acde999925574729be5a3681b70b13c016e55f37620",
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
		want: "adfb4309e020f706ff9ba8f9f035a41e49fecfa061f7230623675c9482a21369",
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
		in      map[v1alpha1.IngressVisibility]sets.String
		want    map[string]sets.String
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
		in: map[v1alpha1.IngressVisibility]sets.String{
			v1alpha1.IngressVisibilityExternalIP:   sets.NewString("foo"),
			v1alpha1.IngressVisibilityClusterLocal: sets.NewString("bar", "baz"),
		},
		want: map[string]sets.String{
			"foo": sets.NewString(
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
		in: map[v1alpha1.IngressVisibility]sets.String{
			v1alpha1.IngressVisibilityExternalIP:   sets.NewString("foo"),
			v1alpha1.IngressVisibilityClusterLocal: sets.NewString("bar", "baz"),
		},
		want: map[string]sets.String{
			"bar": sets.NewString(
				"foo.bar.svc.cluster.local",
				"foo.bar.svc",
				"foo.bar",
			),
			"baz": sets.NewString(
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
