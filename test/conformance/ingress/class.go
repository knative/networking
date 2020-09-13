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

package ingress

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/test"
	"knative.dev/pkg/reconciler"
)

// TestIngressClass verifies that kingrress does not picck ingress up when ingress.class annotation is incorrect.
func TestIngressClass(t *testing.T) {
	t.Parallel()
	ctx, clients := context.Background(), test.Setup(t)

	// Create a backend service to create valid ingress except for invalid ingress.class.
	name, port, _ := CreateRuntimeService(ctx, t, clients, networking.ServicePortNameHTTP1)

	tests := []struct {
		name  string
		class map[string]string
	}{{
		name:  "ommited",
		class: map[string]string{},
	}, {
		name:  "incorrect",
		class: map[string]string{networking.IngressClassAnnotationKey: "incorrect"},
	}, {
		name:  "empty",
		class: map[string]string{networking.IngressClassAnnotationKey: ""},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			createIngressWithClass(ctx, t, clients, test.class, name, port)
		})
	}

}

func createIngressWithClass(ctx context.Context, t *testing.T, clients *test.Clients, class map[string]string, name string, port int) {
	t.Helper()

	org, cancel := CreateIngress(ctx, t, clients, v1alpha1.IngressSpec{
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
	},
		// Override ingress.class annotation
		func(ing *v1alpha1.Ingress) {
			ing.Annotations = class
		},
	)

	const (
		interval = 2 * time.Second
		duration = 6 * time.Second
	)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			var ing *v1alpha1.Ingress
			err := reconciler.RetryTestErrors(func(attempts int) (err error) {
				ing, err = clients.NetworkingClient.Ingresses.Get(ctx, org.Name, metav1.GetOptions{})
				return err
			})
			if err != nil {
				cancel()
				t.Fatal("Error getting Ingress:", err)
			}
			// Verify ingress is not changed.
			if !equality.Semantic.DeepEqual(org, ing) {
				t.Fatalf("Unexpected update, want=%v, got=%v", org, ing)
			}
		case <-time.After(duration):
			break
		}
		break
	}
}
