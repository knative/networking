/*
Copyright 2019 The Knative Authors.

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

package status

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/networking/pkg/http/header"
	"knative.dev/networking/pkg/http/probe"
	"knative.dev/networking/pkg/ingress"

	"go.uber.org/zap/zaptest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ingTemplate = &v1alpha1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "default",
		Name:      "whatever",
	},
	Spec: v1alpha1.IngressSpec{
		Rules: []v1alpha1.IngressRule{{
			Hosts: []string{
				"foo.bar.com",
			},
			Visibility: v1alpha1.IngressVisibilityExternalIP,
			HTTP:       &v1alpha1.HTTPIngressRuleValue{},
		}},
	},
}

func TestProbeAllHosts(t *testing.T) {
	const hostA = "foo.bar.com"
	const hostB = "ksvc.test.dev"
	var hostBEnabled atomic.Bool

	ing := ingTemplate.DeepCopy()
	ing.Spec.Rules[0].Hosts = append(ing.Spec.Rules[0].Hosts, hostB)
	hash, err := ingress.InsertProbe(ing.DeepCopy())
	if err != nil {
		t.Fatal("Failed to insert probe:", err)
	}

	// Failing handler returning HTTP 500 (it should never be called during probing)
	failedRequests := make(chan *http.Request)
	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failedRequests <- r
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Actual probe handler used in Activator and Queue-Proxy
	probeHandler := probe.NewHandler(failHandler)

	// Probes to hostA always succeed and probes to hostB only succeed if hostBEnabled is true
	probeRequests := make(chan *http.Request)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		probeRequests <- r
		if !strings.HasPrefix(r.Host, hostA) &&
			(!hostBEnabled.Load() || !strings.HasPrefix(r.Host, hostB)) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		r.Header.Set(header.HashKey, hash)
		probeHandler.ServeHTTP(w, r)
	})

	ts := httptest.NewServer(finalHandler)
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", ts.URL, err)
	}
	port, err := strconv.Atoi(tsURL.Port())
	if err != nil {
		t.Fatalf("Failed to parse port %q: %v", tsURL.Port(), err)
	}
	hostname := tsURL.Hostname()

	ready := make(chan *v1alpha1.Ingress)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		fakeProbeTargetLister{{
			PodIPs:  sets.New(hostname),
			PodPort: strconv.Itoa(port),
			URLs:    []*url.URL{tsURL},
		}},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	done := make(chan struct{})
	cancelled := prober.Start(done)
	defer func() {
		close(done)
		<-cancelled
	}()

	// The first call to IsReady must succeed and return false
	ok, err := prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	// Wait for both hosts to be probed
	hostASeen, hostBSeen := false, false
	for req := range probeRequests {
		switch req.Host {
		case hostA:
			hostASeen = true
		case hostB:
			hostBSeen = true
		default:
			t.Fatalf("Host header = %q, want %q or %q", req.Host, hostA, hostB)
		}

		if hostASeen && hostBSeen {
			break
		}
	}

	select {
	case <-ready:
		// Since HostB doesn't return 200, the prober shouldn't be ready
		t.Fatal("Prober shouldn't be ready")
	case <-time.After(1 * time.Second):
		// Not ideal but it gives time to the prober to write to ready
		break
	}

	// Make probes to hostB succeed
	hostBEnabled.Store(true)

	// Just drain the requests in the channel to not block the handler
	go func() {
		for range probeRequests {
		}
	}()

	select {
	case <-ready:
		// Wait for the probing to eventually succeed
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}
}

func TestProbeLifecycle(t *testing.T) {
	ing := ingTemplate.DeepCopy()
	hash, err := ingress.InsertProbe(ing.DeepCopy())
	if err != nil {
		t.Fatal("Failed to insert probe:", err)
	}

	// Simulate that the latest configuration is not applied yet by returning a different
	// hash once and then the by returning the expected hash.
	hashes := make(chan string, 1)
	hashes <- "not-the-hash-you-are-looking-for"
	go func() {
		for {
			hashes <- hash
		}
	}()

	// Failing handler returning HTTP 500 (it should never be called during probing)
	failedRequests := make(chan *http.Request)
	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failedRequests <- r
		w.WriteHeader(http.StatusInternalServerError)
	})

	// Actual probe handler used in Activator and Queue-Proxy
	probeHandler := probe.NewHandler(failHandler)

	// Test handler keeping track of received requests, mimicking AppendHeader of K-Network-Hash
	// and simulate a non-existing host by returning 404.
	probeRequests := make(chan *http.Request)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Host, "foo.bar.com") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		probeRequests <- r
		r.Header.Set(header.HashKey, <-hashes)
		probeHandler.ServeHTTP(w, r)
	})

	ts := httptest.NewServer(finalHandler)
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", ts.URL, err)
	}
	port, err := strconv.Atoi(tsURL.Port())
	if err != nil {
		t.Fatalf("Failed to parse port %q: %v", tsURL.Port(), err)
	}
	hostname := tsURL.Hostname()

	ready := make(chan *v1alpha1.Ingress)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		fakeProbeTargetLister{{
			PodIPs:  sets.New(hostname),
			PodPort: strconv.Itoa(port),
			URLs:    []*url.URL{tsURL},
		}},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	done := make(chan struct{})
	cancelled := prober.Start(done)
	defer func() {
		close(done)
		<-cancelled
	}()

	// The first call to IsReady must succeed and return false
	ok, err := prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	const expHostHeader = "foo.bar.com"

	// Wait for the first (failing) and second (success) requests to be executed and validate Host header
	for range 2 {
		req := <-probeRequests
		if req.Host != expHostHeader {
			t.Fatalf("Host header = %q, want %q", req.Host, expHostHeader)
		}
	}

	select {
	case <-ready:
		// Wait for the probing to eventually succeed
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}

	// The subsequent calls to IsReady must succeed and return true
	for range 5 {
		if ok, err = prober.IsReady(context.Background(), ing); err != nil {
			t.Fatal("IsReady failed:", err)
		}
		if !ok {
			t.Fatal("IsReady() returned false")
		}
	}

	// Cancel Ingress probing -> deletes the cached state
	prober.CancelIngressProbing(ing)

	select {
	// Validate that no probe requests were issued (cached)
	case <-probeRequests:
		t.Fatal("An unexpected probe request was received")
	// Validate that no requests went through the probe handler
	case <-failedRequests:
		t.Fatal("An unexpected request went through the probe handler")
	default:
	}

	// The state has been removed and IsReady must return False
	ok, err = prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	// Wait for the first request (success) to be executed
	<-probeRequests

	select {
	case <-ready:
		// Wait for the probing to eventually succeed
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}

	select {
	// Validate that no requests went through the probe handler
	case <-failedRequests:
		t.Fatal("An unexpected request went through the probe handler")
	default:
		break
	}
}

func TestProbeListerFail(t *testing.T) {
	ing := ingTemplate.DeepCopy()
	ready := make(chan *v1alpha1.Ingress)
	defer close(ready)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		notFoundLister{},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	// If we can't list, this  must fail and return false
	ok, err := prober.IsReady(context.Background(), ing)
	if err == nil {
		t.Fatal("IsReady returned unexpected success")
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}
}

func TestCancelPodProbing(t *testing.T) {
	type timedRequest struct {
		*http.Request
		Time time.Time
	}

	// Handler keeping track of received requests and mimicking an Ingress not ready
	requests := make(chan *timedRequest, 100)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- &timedRequest{
			Time:    time.Now(),
			Request: r,
		}
		w.WriteHeader(http.StatusNotFound)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", ts.URL, err)
	}
	port, err := strconv.Atoi(tsURL.Port())
	if err != nil {
		t.Fatalf("Failed to parse port %q: %v", tsURL.Port(), err)
	}
	hostname := tsURL.Hostname()
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "gateway",
		},
		Status: v1.PodStatus{
			PodIP: strings.Split(tsURL.Host, ":")[0],
		},
	}

	ready := make(chan *v1alpha1.Ingress)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		fakeProbeTargetLister{{
			PodIPs:  sets.New(hostname),
			PodPort: strconv.Itoa(port),
			URLs:    []*url.URL{tsURL},
		}},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	done := make(chan struct{})
	cancelled := prober.Start(done)
	defer func() {
		close(done)
		<-cancelled
	}()

	ing := ingTemplate.DeepCopy()
	ok, err := prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	select {
	case <-requests:
		// Wait for the first probe request
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}

	// Create a new version of the Ingress (to replace the original Ingress)
	const otherDomain = "blabla.net"
	ing = ing.DeepCopy()
	ing.Spec.Rules[0].Hosts[0] = otherDomain

	// Create a different Ingress (to be probed in parallel)
	const parallelDomain = "parallel.net"
	func() {
		dc := ing.DeepCopy()
		dc.Spec.Rules[0].Hosts[0] = parallelDomain
		dc.Name = "something"

		ok, err = prober.IsReady(context.Background(), dc)
		if err != nil {
			t.Fatal("IsReady failed:", err)
		}
		if ok {
			t.Fatal("IsReady() returned true")
		}
	}()

	// Check that probing is unsuccessful
	select {
	case <-ready:
		t.Fatal("Probing succeeded while it should not have succeeded")
	default:
	}

	ok, err = prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	// Drain requests for the old version
	for req := range requests {
		t.Log("req.Host:", req.Host)
		if strings.HasPrefix(req.Host, otherDomain) {
			break
		}
	}

	// Cancel Pod probing
	prober.CancelPodProbing(pod)
	cancelTime := time.Now()

	// Check that there are no requests for the old Ingress and the requests predate cancellation
	for {
		select {
		case req := <-requests:
			if !strings.HasPrefix(req.Host, otherDomain) &&
				!strings.HasPrefix(req.Host, parallelDomain) {
				t.Fatalf("Host = %s, want: %s or %s", req.Host, otherDomain, parallelDomain)
			} else if req.Time.Sub(cancelTime) > 0 {
				t.Fatal("Request was made after cancellation")
			}
		default:
			return
		}
	}
}

func TestPartialPodCancellation(t *testing.T) {
	ing := ingTemplate.DeepCopy()
	hash, err := ingress.InsertProbe(ing.DeepCopy())
	if err != nil {
		t.Fatal("Failed to insert probe:", err)
	}

	// Simulate a probe target returning HTTP 200 OK and the correct hash
	requests := make(chan *http.Request, 100)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.Header().Set(header.HashKey, hash)
		w.WriteHeader(http.StatusOK)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", ts.URL, err)
	}
	port, err := strconv.Atoi(tsURL.Port())
	if err != nil {
		t.Fatalf("Failed to parse port %q: %v", tsURL.Port(), err)
	}

	// pods[0] will be probed successfully, pods[1] will never be probed successfully
	pods := []*v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod0",
		},
		Status: v1.PodStatus{
			PodIP: strings.Split(tsURL.Host, ":")[0],
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "pod1",
		},
		Status: v1.PodStatus{
			PodIP: "198.51.100.1",
		},
	}}

	ready := make(chan *v1alpha1.Ingress)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		fakeProbeTargetLister{{
			PodIPs:  sets.New(pods[0].Status.PodIP, pods[1].Status.PodIP),
			PodPort: strconv.Itoa(port),
			URLs:    []*url.URL{tsURL},
		}},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	done := make(chan struct{})
	cancelled := prober.Start(done)
	defer func() {
		close(done)
		<-cancelled
	}()

	ok, err := prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	select {
	case <-requests:
		// Wait for the first probe request
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}

	// Check that probing is unsuccessful
	select {
	case <-ready:
		t.Fatal("Probing succeeded while it should not have succeeded")
	default:
	}

	// Cancel probing of pods[1]
	prober.CancelPodProbing(pods[1])

	// Check that probing was successful
	select {
	case <-ready:
		break
	case <-time.After(5 * time.Second):
		t.Fatal("Probing was not successful even after waiting")
	}
}

func TestCancelIngressProbing(t *testing.T) {
	ing := ingTemplate.DeepCopy()
	// Handler keeping track of received requests and mimicking an Ingress not ready
	requests := make(chan *http.Request, 100)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.WriteHeader(http.StatusNotFound)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", ts.URL, err)
	}
	port, err := strconv.Atoi(tsURL.Port())
	if err != nil {
		t.Fatalf("Failed to parse port %q: %v", tsURL.Port(), err)
	}
	hostname := tsURL.Hostname()

	ready := make(chan *v1alpha1.Ingress)
	prober := NewProber(
		zaptest.NewLogger(t).Sugar(),
		fakeProbeTargetLister{{
			PodIPs:  sets.New(hostname),
			PodPort: strconv.Itoa(port),
			URLs:    []*url.URL{tsURL},
		}},
		func(ing *v1alpha1.Ingress) {
			ready <- ing
		})

	done := make(chan struct{})
	cancelled := prober.Start(done)
	defer func() {
		close(done)
		<-cancelled
	}()

	ok, err := prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	select {
	case <-requests:
		// Wait for the first probe request
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for probing to succeed.")
	}

	const domain = "blabla.net"

	// Create a new version of the Ingress
	ing = ing.DeepCopy()
	ing.Spec.Rules[0].Hosts[0] = domain

	// Check that probing is unsuccessful
	select {
	case <-ready:
		t.Fatal("Probing succeeded while it should not have succeeded")
	default:
	}

	ok, err = prober.IsReady(context.Background(), ing)
	if err != nil {
		t.Fatal("IsReady failed:", err)
	}
	if ok {
		t.Fatal("IsReady() returned true")
	}

	// Drain requests for the old version.
	for req := range requests {
		t.Log("req.Host:", req.Host)
		if strings.HasPrefix(req.Host, domain) {
			break
		}
	}

	// Cancel Ingress probing.
	prober.CancelIngressProbing(ing)

	// Check that the requests were for the new version.
	close(requests)
	for req := range requests {
		if !strings.HasPrefix(req.Host, domain) {
			t.Fatalf("Host = %s, want: %s", req.Host, domain)
		}
	}
}

func TestProbeVerifier(t *testing.T) {
	const hash = "Hi! I am hash!"
	prober := NewProber(zaptest.NewLogger(t).Sugar(), nil, nil)
	verifier := prober.probeVerifier(&workItem{
		ingressState: &ingressState{
			ing: &v1alpha1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "bar",
				},
			},
			hash: hash,
		},
		podState: nil,
		context:  nil,
		url:      nil,
		podIP:    "",
		podPort:  "",
		logger:   zaptest.NewLogger(t).Sugar(),
	})
	cases := []struct {
		name string
		resp *http.Response
		want bool
	}{{
		name: "HTTP 200 matching hash",
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{header.HashKey: []string{hash}},
		},
		want: true,
	}, {
		name: "HTTP 200 mismatching hash",
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{header.HashKey: []string{"nope"}},
		},
		want: false,
	}, {
		name: "HTTP 200 missing header",
		resp: &http.Response{
			StatusCode: http.StatusOK,
		},
		want: true,
	}, {
		name: "HTTP 404",
		resp: &http.Response{
			StatusCode: http.StatusNotFound,
		},
		want: false,
	}, {
		name: "HTTP 503",
		resp: &http.Response{
			StatusCode: http.StatusServiceUnavailable,
		},
		want: false,
	}, {
		name: "HTTP 403",
		resp: &http.Response{
			StatusCode: http.StatusForbidden,
		},
		want: true,
	}, {
		name: "HTTP 503",
		resp: &http.Response{
			StatusCode: http.StatusServiceUnavailable,
		},
		want: false,
	}, {
		name: "HTTP 301",
		resp: &http.Response{
			StatusCode: http.StatusMovedPermanently,
		},
		want: true,
	}, {
		name: "HTTP 302",
		resp: &http.Response{
			StatusCode: http.StatusFound,
		},
		want: true,
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := verifier(c.resp, nil)
			if got != c.want {
				t.Errorf("got: %v, want: %v", got, c.want)
			}
		})
	}
}

type fakeProbeTargetLister []ProbeTarget

func (l fakeProbeTargetLister) ListProbeTargets(_ context.Context, ing *v1alpha1.Ingress) ([]ProbeTarget, error) {
	targets := []ProbeTarget{}
	for _, target := range l {
		newTarget := ProbeTarget{
			PodIPs:  target.PodIPs,
			PodPort: target.PodPort,
			Port:    target.Port,
		}
		for _, url := range target.URLs {
			for _, host := range ing.Spec.Rules[0].Hosts {
				newURL := *url
				newURL.Host = host
				newTarget.URLs = append(newTarget.URLs, &newURL)
			}
		}
		targets = append(targets, newTarget)
	}
	return targets, nil
}

type notFoundLister struct{}

func (l notFoundLister) ListProbeTargets(_ context.Context, _ *v1alpha1.Ingress) ([]ProbeTarget, error) {
	return nil, errors.New("not found")
}
