/*
Copyright 2020 The Knative Authors.

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
	"testing"

	"github.com/google/go-cmp/cmp"
	"knative.dev/pkg/apis"
)

func TestDomainGetConditionSet(t *testing.T) {
	d := Domain{}

	if got, want := d.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetConditionSet=%v, want=%v", got, want)
	}
}

func TestDomainGetGroupVersionKind(t *testing.T) {
	d := Domain{}
	expected := SchemeGroupVersion.WithKind("Domain")
	if !cmp.Equal(expected, d.GetGroupVersionKind()) {
		t.Error("Unexpected diff (-want, +got) =", cmp.Diff(expected, d.GetGroupVersionKind()))
	}
}
