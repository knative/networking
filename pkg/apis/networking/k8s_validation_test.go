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

package networking

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
)

func TestObjectReferenceValidation(t *testing.T) {
	tests := []struct {
		name string
		r    *corev1.ObjectReference
		want *apis.FieldError
	}{{
		name: "nil",
	}, {
		name: "no api version",
		r: &corev1.ObjectReference{
			Kind: "Bar",
			Name: "foo",
		},
		want: apis.ErrMissingField("apiVersion"),
	}, {
		name: "bad api version",
		r: &corev1.ObjectReference{
			APIVersion: "/v1alpha1",
			Kind:       "Bar",
			Name:       "foo",
		},
		want: apis.ErrInvalidValue("prefix part must be non-empty", "apiVersion"),
	}, {
		name: "no kind",
		r: &corev1.ObjectReference{
			APIVersion: "foo/v1alpha1",
			Name:       "foo",
		},
		want: apis.ErrMissingField("kind"),
	}, {
		name: "bad kind",
		r: &corev1.ObjectReference{
			APIVersion: "foo/v1alpha1",
			Kind:       "Bad Kind",
			Name:       "foo",
		},
		want: apis.ErrInvalidValue("a valid C identifier must start with alphabetic character or '_', followed by a string of alphanumeric characters or '_' (e.g. 'my_name',  or 'MY_NAME',  or 'MyName', regex used for validation is '[A-Za-z_][A-Za-z0-9_]*')", "kind"),
	}, {
		name: "no namespace",
		r: &corev1.ObjectReference{
			APIVersion: "foo.group/v1alpha1",
			Kind:       "Bar",
			Name:       "the-bar-0001",
		},
		want: nil,
	}, {
		name: "no name",
		r: &corev1.ObjectReference{
			APIVersion: "foo.group/v1alpha1",
			Kind:       "Bar",
		},
		want: apis.ErrMissingField("name"),
	}, {
		name: "bad name",
		r: &corev1.ObjectReference{
			APIVersion: "foo.group/v1alpha1",
			Kind:       "Bar",
			Name:       "bad name",
		},
		want: apis.ErrInvalidValue("a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')", "name"),
	}, {
		name: "disallowed fields",
		r: &corev1.ObjectReference{
			APIVersion: "foo.group/v1alpha1",
			Kind:       "Bar",
			Name:       "bar0001",

			// None of these are allowed.
			Namespace:       "foo",
			FieldPath:       "some.field.path",
			ResourceVersion: "234234",
			UID:             "deadbeefcafebabe",
		},
		want: apis.ErrDisallowedFields("namespace", "fieldPath", "resourceVersion", "uid"),
	}, {
		name: "all good",
		r: &corev1.ObjectReference{
			APIVersion: "foo.group/v1alpha1",
			Kind:       "Bar",
			Name:       "bar0001",
		},
		want: nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ValidateNamespacedObjectReference(test.r)
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("ValidateNamespacedObjectReference (-want, +got): \n%s", diff)
			}
		})
	}
}
