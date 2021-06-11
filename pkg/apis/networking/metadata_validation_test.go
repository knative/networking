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

	"knative.dev/pkg/apis"
)

func TestValidateObjectMetadata(t *testing.T) {
	cases := []struct {
		name        string
		annotations map[string]string
		expectErr   *apis.FieldError
	}{{
		name: "invalid knative prefix annotation",
		annotations: map[string]string{
			"networking.knative.dev/testAnnotation": "value",
		},
		expectErr: apis.ErrInvalidKeyName("networking.knative.dev/testAnnotation", apis.CurrentField),
	}, {
		name: "valid non-knative prefix annotation key",
		annotations: map[string]string{
			"testAnnotation": "testValue",
		},
	}, {
		name: "valid disable auto TLS annotation key",
		annotations: map[string]string{
			DisableAutoTLSAnnotationKey: "true",
		},
	}, {
		name: "valid certificate class annotation key",
		annotations: map[string]string{
			CertificateClassAnnotationKey: "certificate-class",
		},
	}, {
		name: "valid http option annotation key",
		annotations: map[string]string{
			HTTPOptionAnnotationKey: "Redirected",
		},
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateAnnotations(c.annotations)
			if got, want := err.Error(), c.expectErr.Error(); got != want {
				t.Errorf("Got: %q, want: %q", got, want)
			}
		})
	}
}
