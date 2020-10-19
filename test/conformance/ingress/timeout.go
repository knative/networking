/*
Copyright 2019 The Knative Authors

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
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/test"
)

// TestTimeout verifies that an Ingress implements "no timeout".
func TestTimeout(t *test.T) {
	t.Parallel()

	name, port, _ := CreateTimeoutService(t.C, t, t.Clients)

	// Create a simple Ingress over the Service.
	_, client, _ := CreateIngressReady(t.C, t, t.Clients, v1alpha1.IngressSpec{
		Rules: []v1alpha1.IngressRule{{
			Hosts:      []string{name + ".example.com"},
			Visibility: v1alpha1.IngressVisibilityExternalIP,
			HTTP: &v1alpha1.HTTPIngressRuleValue{
				Paths: []v1alpha1.HTTPIngressPath{{
					Splits: []v1alpha1.IngressBackendSplit{{
						IngressBackend: v1alpha1.IngressBackend{
							ServiceName:      name,
							ServiceNamespace: test.ServingNamespace,
							ServicePort:      intstr.FromInt(port),
						},
					}},
				}},
			},
		}},
	})

	const timeout = 10 * time.Second

	cases := []struct {
		name         string
		code         int
		initialDelay time.Duration
		delay        time.Duration
	}{{
		name: "no delays is OK",
		code: http.StatusOK,
	}, {
		name:         "large delay before headers is ok",
		code:         http.StatusOK,
		initialDelay: timeout,
	}, {
		name:  "large delay after headers is ok",
		code:  http.StatusOK,
		delay: timeout,
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *test.T) {
			checkTimeout(t.C, t, client, name, c.code, c.initialDelay, c.delay)
		})
	}
}

func checkTimeout(ctx context.Context, t *test.T, client *http.Client, name string, code int, initial time.Duration, timeout time.Duration) {
	t.Helper()

	resp, err := client.Get(fmt.Sprintf("http://%s.example.com?initialTimeout=%d&timeout=%d",
		name, initial.Milliseconds(), timeout.Milliseconds()))
	if err != nil {
		t.Fatal("Error making GET request:", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != code {
		t.Errorf("Unexpected status code: %d, wanted %d", resp.StatusCode, code)
		DumpResponse(ctx, t, resp)
	}
}
