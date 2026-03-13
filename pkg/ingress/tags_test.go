/*
Copyright 2026 The Knative Authors

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
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apisnet "knative.dev/networking/pkg/apis/networking"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/pkg/http/header"
)

func TestTagToHosts(t *testing.T) {
	tests := []struct {
		name string
		ing  *v1alpha1.Ingress
		want map[string]sets.Set[string]
	}{{
		name: "missing annotation",
		ing:  &v1alpha1.Ingress{},
	}, {
		name: "invalid annotation",
		ing: &v1alpha1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				apisnet.TagToHostAnnotationKey: "{",
			}},
		},
	}, {
		name: "valid annotation",
		ing: &v1alpha1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				apisnet.TagToHostAnnotationKey: `{"blue":["blue.example.com","blue.example.com"],"green":[],"internal":["green.test-ns.svc.cluster.local"]}`,
			}},
		},
		want: map[string]sets.Set[string]{
			"blue":     sets.New("blue.example.com"),
			"internal": sets.New("green.test-ns.svc.cluster.local"),
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(asSortedSlices(test.want), asSortedSlices(TagToHosts(test.ing))); diff != "" {
				t.Fatalf("TagToHosts diff (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestHostsForTag(t *testing.T) {
	tagToHosts := map[string]sets.Set[string]{
		"blue": sets.New(
			"blue.example.com",
			"blue.test-ns.svc.cluster.local",
		),
	}

	if diff := cmp.Diff(
		sets.List(sets.New("blue.example.com")),
		sets.List(HostsForTag("blue", v1alpha1.IngressVisibilityExternalIP, tagToHosts)),
	); diff != "" {
		t.Fatalf("external HostsForTag diff (-want,+got):\n%s", diff)
	}

	if diff := cmp.Diff(
		sets.List(sets.New(
			"blue.test-ns",
			"blue.test-ns.svc",
			"blue.test-ns.svc.cluster.local",
		)),
		sets.List(HostsForTag("blue", v1alpha1.IngressVisibilityClusterLocal, tagToHosts)),
	); diff != "" {
		t.Fatalf("cluster-local HostsForTag diff (-want,+got):\n%s", diff)
	}
}

func TestMakeTagHostIngressPath(t *testing.T) {
	original := &v1alpha1.HTTPIngressPath{
		Headers: map[string]v1alpha1.HeaderMatch{
			header.RouteTagKey: {Exact: "blue"},
			"X-Test":           {Exact: "preserved"},
		},
		AppendHeaders: map[string]string{
			"X-Existing": "value",
		},
	}

	got := MakeTagHostIngressPath(original, "blue")

	if diff := cmp.Diff(
		map[string]string{
			"X-Test": headerMatchValue(got.Headers["X-Test"]),
		},
		map[string]string{
			"X-Test": "preserved",
		},
	); diff != "" {
		t.Fatalf("MakeTagHostIngressPath headers diff (-want,+got):\n%s", diff)
	}

	if got.AppendHeaders[header.RouteTagKey] != "blue" {
		t.Fatalf("MakeTagHostIngressPath append tag = %q, want %q", got.AppendHeaders[header.RouteTagKey], "blue")
	}
	if _, ok := original.AppendHeaders[header.RouteTagKey]; ok {
		t.Fatal("MakeTagHostIngressPath mutated original append headers")
	}
	if _, ok := got.Headers[header.RouteTagKey]; ok {
		t.Fatal("MakeTagHostIngressPath kept route tag header match")
	}
}

func TestRouteHosts(t *testing.T) {
	ruleHosts := sets.New("route.example.com")
	tagToHosts := map[string]sets.Set[string]{
		"blue": sets.New("blue.example.com"),
	}

	hostPath := &v1alpha1.HTTPIngressPath{
		AppendHeaders: map[string]string{
			header.RouteTagKey: "blue",
		},
	}
	if diff := cmp.Diff(
		sets.List(sets.New("blue.example.com", "route.example.com")),
		sets.List(RouteHosts(ruleHosts, hostPath, v1alpha1.IngressVisibilityExternalIP, tagToHosts)),
	); diff != "" {
		t.Fatalf("host RouteHosts diff (-want,+got):\n%s", diff)
	}

	headerPath := &v1alpha1.HTTPIngressPath{
		Headers: map[string]v1alpha1.HeaderMatch{
			header.RouteTagKey: {Exact: "blue"},
		},
		AppendHeaders: map[string]string{
			header.RouteTagKey: "blue",
		},
	}
	if diff := cmp.Diff(
		sets.List(ruleHosts),
		sets.List(RouteHosts(ruleHosts, headerPath, v1alpha1.IngressVisibilityExternalIP, tagToHosts)),
	); diff != "" {
		t.Fatalf("header RouteHosts diff (-want,+got):\n%s", diff)
	}
}

func TestHostRouteTags(t *testing.T) {
	rule := &v1alpha1.IngressRule{
		HTTP: &v1alpha1.HTTPIngressRuleValue{
			Paths: []v1alpha1.HTTPIngressPath{{
				AppendHeaders: map[string]string{
					header.RouteTagKey: "blue",
				},
			}, {
				Headers: map[string]v1alpha1.HeaderMatch{
					header.RouteTagKey: {Exact: "green"},
				},
			}, {
				Headers: map[string]v1alpha1.HeaderMatch{
					header.RouteTagKey: {Exact: "red"},
				},
				AppendHeaders: map[string]string{
					header.RouteTagKey: "red",
				},
			}},
		},
	}

	if diff := cmp.Diff(sets.List(sets.New("blue")), sets.List(HostRouteTags(rule))); diff != "" {
		t.Fatalf("HostRouteTags diff (-want,+got):\n%s", diff)
	}
}

func asSortedSlices(in map[string]sets.Set[string]) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for tag, hosts := range in {
		out[tag] = sets.List(hosts)
	}
	return out
}

func headerMatchValue(match v1alpha1.HeaderMatch) string {
	return match.Exact
}
