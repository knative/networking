/*
Copyright 2022 The Knative Authors

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

package http

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIsPotentialMeshErrorResponse(t *testing.T) {
	for _, test := range []struct {
		statusCode int
		expect     bool
	}{{
		statusCode: 404,
		expect:     false,
	}, {
		statusCode: 200,
		expect:     false,
	}, {
		statusCode: 501,
		expect:     false,
	}, {
		statusCode: 502,
		expect:     true,
	}, {
		statusCode: 503,
		expect:     true,
	}} {
		t.Run(fmt.Sprintf("statusCode=%d", test.statusCode), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: test.statusCode,
			}
			if got := IsPotentialMeshErrorResponse(resp); got != test.expect {
				t.Errorf("IsPotentialMeshErrorResponse({StatusCode: %d}) = %v, expected %v", resp.StatusCode, got, test.expect)
			}
		})
	}
}
