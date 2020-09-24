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
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/test"
	ping "knative.dev/networking/test/test_images/grpc-ping/proto"
	"knative.dev/pkg/test/logstream"
)

// TestHTTP1AndEmptyPort verifies that an empty port name uses HTTP1. This should be the current behavior.
func TestHTTP1AndEmptyPort(t *testing.T) {
	ctx, clients := context.Background(), test.Setup(t)
	name, port, _ := CreateRuntimeService(ctx, t, clients, "")

	// Create a simple Ingress over the Service.
	_, client, _ := CreateIngressReady(ctx, t, clients, v1alpha1.IngressSpec{
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

	ri := RuntimeRequest(ctx, t, client, "http://"+name+".example.com")
	if ri == nil {
		return
	}

	if want, got := 1, ri.Request.ProtoMajor; want != got {
		t.Errorf("ProtoMajor = %d, wanted %d", got, want)
	}
}

// TestHTTP2AndEmptyPort verifies that an empty port name uses HTTP2.
// This is not the current behavior.
func TestHTTP2AndEmptyPort(t *testing.T) {
	ctx, clients := context.Background(), test.Setup(t)
	name, port, _ := CreateRuntimeService(ctx, t, clients, "")

	// Create a simple Ingress over the Service.
	_, client, _ := CreateIngressReady(ctx, t, clients, v1alpha1.IngressSpec{
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

	ri := RuntimeRequest(ctx, t, client, "http://"+name+".example.com")
	if ri == nil {
		return
	}

	if want, got := 2, ri.Request.ProtoMajor; want != got {
		t.Errorf("ProtoMajor = %d, wanted %d", got, want)
	}
}

// TestGRPCWithEmptyPort verifies that a nameless port can establish a basic GRPC connection.
func TestGRPCWithEmptyPort(t *testing.T) {
	t.Parallel()
	defer logstream.Start(t)()
	ctx, clients := context.Background(), test.Setup(t)

	const suffix = "- pong"
	name, port, _ := CreateGRPCServiceWithPortName(ctx, t, clients, suffix, "")

	domain := name + ".example.com"

	// Create a simple Ingress over the Service.
	_, dialCtx, _ := createIngressReadyDialContext(ctx, t, clients, v1alpha1.IngressSpec{
		Rules: []v1alpha1.IngressRule{{
			Hosts:      []string{domain},
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

	conn, err := grpc.Dial(
		domain+":80",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return dialCtx(ctx, "unused", addr)
		}),
	)
	if err != nil {
		t.Fatal("Dial() =", err)
	}
	defer conn.Close()
	pc := ping.NewPingServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := pc.PingStream(ctx)
	if err != nil {
		t.Fatal("PingStream() =", err)
	}

	for i := 0; i < 100; i++ {
		checkGRPCRoundTrip(t, stream, suffix)
	}
}
