/*
Copyright 2020 The Knative Authors

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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func TestDomainSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		ds   DomainSpec
		want *apis.FieldError
	}{{
		name: "all good",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
			LoadBalancers: []LoadBalancerIngressSpec{{
				Domain: "test-domain",
			}},
		},
	}, {
		name: "no spec",
		ds:   DomainSpec{},
		want: apis.ErrMissingField("spec"),
	}, {
		name: "ingress class isnt specified",
		ds: DomainSpec{
			LoadBalancers: []LoadBalancerIngressSpec{{
				Domain: "test-domain",
			}},
		},
		want: apis.ErrMissingField("spec.ingressClass"),
	}, {
		name: "loadbalacers arent specified",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
		},
		want: apis.ErrMissingField("spec.loadBalancers"),
	}, {
		name: "at least one field in loadbalacer is specified",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
			LoadBalancers: []LoadBalancerIngressSpec{{
				Domain: "some-domain",
			}},
		},
	}, {
		name: "none of the fields are specified in a loadbalancer",
		ds: DomainSpec{
			IngressClass:  "test-ingress-class",
			LoadBalancers: []LoadBalancerIngressSpec{{}},
		},
		want: apis.ErrMissingOneOf("spec.loadBalancers[0].domain", "spec.loadBalancers[0].domainInternal",
			"spec.loadBalancers[0].ip", "spec.loadBalancers[0].meshOnly"),
	}, {
		name: "name is missing from ingressConfig",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
			Configs:      []IngressConfig{{Namespace: "ns", Type: "my-type"}},
			LoadBalancers: []LoadBalancerIngressSpec{{
				Domain: "some-domain",
			}},
		},
		want: apis.ErrMissingField("spec.configs[0].name"),
	}, {
		name: "type is missing from ingressConfig",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
			Configs:      []IngressConfig{{Namespace: "ns", Name: "my-name"}},
			LoadBalancers: []LoadBalancerIngressSpec{{
				Domain: "some-domain",
			}},
		},
		want: apis.ErrMissingField("spec.configs[0].type"),
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ds := Domain{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ds"},
				Spec:       test.ds,
			}
			got := ds.Validate(context.Background())
			if !cmp.Equal(test.want.Error(), got.Error()) {
				t.Errorf("Validate (-want, +got) = \n%s", cmp.Diff(test.want.Error(), got.Error()))
			}
		})
	}
}
