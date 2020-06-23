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

func TestRealmSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		rs   RealmSpec
		want *apis.FieldError
	}{{
		name: "external domain is specified",
		rs: RealmSpec{
			External: "test-ext",
		},
	}, {
		name: "cluster domain is specified",
		rs: RealmSpec{
			Cluster: "test-cluster",
		},
	}, {
		name: "neither cluster nor external domain is specified",
		rs:   RealmSpec{},
		want: apis.ErrMissingOneOf("spec.cluster", "spec.external"),
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := Realm{
				ObjectMeta: metav1.ObjectMeta{Name: "test-realm"},
				Spec:       test.rs,
			}
			got := rs.Validate(context.Background())
			if !cmp.Equal(test.want.Error(), got.Error()) {
				t.Errorf("Validate (-want, +got) = \n%s", cmp.Diff(test.want.Error(), got.Error()))
			}
		})
	}
}
