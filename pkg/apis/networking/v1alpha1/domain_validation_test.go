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
		want: apis.ErrMissingField("spec.IngressClass"),
	}, {
		name: "loadbalacers arent specified",
		ds: DomainSpec{
			IngressClass: "test-ingress-class",
		},
		want: apis.ErrMissingField("spec.LoadBalancers"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ds := Domain{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ds"},
				Spec:       test.ds,
			}
			got := ds.Validate(context.Background())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("Validate (-want, +got) = %v", diff)
			}
		})
	}
}