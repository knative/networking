/*
Copyright 2023 The Knative Authors

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

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	hosts = []string{"foo", "bar", "foo.bar"}
)

func TestGetIngressTLSForVisibility(t *testing.T) {
	tests := []struct {
		name       string
		visibility IngressVisibility
		ingress    *Ingress
		want       []IngressTLS
	}{{
		name:       "no TLS entries",
		visibility: IngressVisibilityClusterLocal,
		ingress: &Ingress{
			Spec: IngressSpec{
				Rules: []IngressRule{
					{
						Hosts:      hosts,
						Visibility: IngressVisibilityClusterLocal,
					},
					{
						Hosts:      []string{"other", "entries"},
						Visibility: IngressVisibilityExternalIP,
					},
				},
				TLS: make([]IngressTLS, 0),
			},
		},
		want: make([]IngressTLS, 0),
	}, {
		name:       "no matching entries",
		visibility: IngressVisibilityClusterLocal,
		ingress: &Ingress{
			Spec: IngressSpec{
				Rules: []IngressRule{
					{
						Hosts:      hosts,
						Visibility: IngressVisibilityClusterLocal,
					},
					{
						Hosts:      []string{"other", "entries"},
						Visibility: IngressVisibilityExternalIP,
					},
				},
				TLS: []IngressTLS{
					{Hosts: []string{"something", "else"}},
				},
			},
		},
		want: make([]IngressTLS, 0),
	}, {
		name:       "matching cluster-local entries",
		visibility: IngressVisibilityClusterLocal,
		ingress: &Ingress{
			Spec: IngressSpec{
				Rules: []IngressRule{
					{
						Hosts:      hosts,
						Visibility: IngressVisibilityClusterLocal,
					},
					{
						Hosts:      []string{"other", "entries"},
						Visibility: IngressVisibilityExternalIP,
					},
				},
				TLS: []IngressTLS{
					{Hosts: hosts},
				},
			},
		},
		want: []IngressTLS{{Hosts: hosts}},
	}, {
		name:       "matching external-ip entries",
		visibility: IngressVisibilityExternalIP,
		ingress: &Ingress{
			Spec: IngressSpec{
				Rules: []IngressRule{
					{
						Hosts:      hosts,
						Visibility: IngressVisibilityExternalIP,
					},
					{
						Hosts:      []string{"other", "entries"},
						Visibility: IngressVisibilityClusterLocal,
					},
				},
				TLS: []IngressTLS{
					{Hosts: hosts},
				},
			},
		},
		want: []IngressTLS{{Hosts: hosts}},
	}, {
		name:       "matching entries with different visibility",
		visibility: IngressVisibilityClusterLocal,
		ingress: &Ingress{
			Spec: IngressSpec{
				Rules: []IngressRule{
					{
						Hosts:      hosts,
						Visibility: IngressVisibilityExternalIP,
					},
					{
						Hosts:      []string{"other", "entries"},
						Visibility: IngressVisibilityClusterLocal,
					},
				},
				TLS: []IngressTLS{
					{Hosts: hosts},
				},
			},
		},
		want: make([]IngressTLS, 0),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ingress.GetIngressTLSForVisibility(test.visibility)

			if !cmp.Equal(test.want, got) {
				t.Errorf("GetIngressTLSForVisibility (-want, +got) = \n%s", cmp.Diff(test.want, got))
			}
		})
	}
}
