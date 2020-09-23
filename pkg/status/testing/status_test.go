package testing

import (
	context "context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func TestCallsFakeIsReady(t *testing.T) {
	calledRightFunction := false
	statusManager := FakeStatusManager{
		FakeIsReady: func(context.Context, *v1alpha1.Ingress) (bool, error) {
			calledRightFunction = true
			return true, nil
		},
	}

	statusManager.IsReady(context.Background(), &v1alpha1.Ingress{})

	if !calledRightFunction {
		t.Errorf("calledRightFunction = false, want true")
	}
}

func TestIsReadyCallCount(t *testing.T) {
	statusManager := FakeStatusManager{
		FakeIsReady: func(context.Context, *v1alpha1.Ingress) (bool, error) {
			return true, nil
		},
	}
	ingress1 := v1alpha1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "name1"}}
	ingress2 := v1alpha1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: "ns2", Name: "name2"}}
	ingress3 := v1alpha1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: "ns3", Name: "name3"}}

	for i := 0; i < 10; i++ {
		statusManager.IsReady(context.Background(), &ingress1)
	}

	for i := 0; i < 5; i++ {
		statusManager.IsReady(context.Background(), &ingress2)
	}

	if statusManager.IsReadyCallCount(&ingress1) != 10 {
		t.Errorf("statusManager.IsReadyCallCount(&ingress1) = %v, want %v",
			statusManager.IsReadyCallCount(&ingress1), 10)
	}

	if statusManager.IsReadyCallCount(&ingress2) != 5 {
		t.Errorf("statusManager.IsReadyCallCount(&ingress2) = %v, want %v",
			statusManager.IsReadyCallCount(&ingress1), 5)
	}

	if statusManager.IsReadyCallCount(&ingress3) != 0 {
		t.Errorf("statusManager.IsReadyCallCount(&ingress3) = %v, want %v",
			statusManager.IsReadyCallCount(&ingress3), 0)
	}
}
