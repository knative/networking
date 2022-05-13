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
package k8s

import (
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNameForPortNumber(t *testing.T) {
	for _, tc := range []struct {
		name       string
		svc        *corev1.Service
		portNumber int32
		portName   string
		err        error
	}{{
		name: "HTTP to 80",
		svc: &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 80,
					Name: "http",
				}, {
					Port: 443,
					Name: "https",
				}},
			},
		},
		portName:   "http",
		portNumber: 80,
	}, {
		name: "no port",
		svc: &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 443,
					Name: "https",
				}},
			},
		},
		portNumber: 80,
		err:        errors.New("no port with number 80 found"),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			portName, err := NameForPortNumber(tc.svc, tc.portNumber)
			if !reflect.DeepEqual(err, tc.err) { // cmp Doesn't work well here due to private fields.
				t.Errorf("Err = %v, want: %v", err, tc.err)
			}
			if tc.err == nil && portName != tc.portName {
				t.Errorf("PortName = %s, want: %s", portName, tc.portName)
			}
		})
	}
}

func TestPortNumberForName(t *testing.T) {
	for _, tc := range []struct {
		name       string
		subset     corev1.EndpointSubset
		portNumber int32
		portName   string
		err        error
	}{{
		name: "HTTP to 80",
		subset: corev1.EndpointSubset{
			Ports: []corev1.EndpointPort{{
				Port: 8080,
				Name: "http",
			}, {
				Port: 8443,
				Name: "https",
			}},
		},
		portName:   "http",
		portNumber: 8080,
	}, {
		name: "no port",
		subset: corev1.EndpointSubset{
			Ports: []corev1.EndpointPort{{
				Port: 8443,
				Name: "https",
			}},
		},
		portName: "http",
		err:      errors.New(`no port for name "http" found`),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			portNumber, err := PortNumberForName(tc.subset, tc.portName)
			if !reflect.DeepEqual(err, tc.err) { // cmp Doesn't work well here due to private fields.
				t.Errorf("Err = %v, want: %v", err, tc.err)
			}
			if tc.err == nil && portNumber != tc.portNumber {
				t.Errorf("PortNumber = %d, want: %d", portNumber, tc.portNumber)
			}
		})
	}
}
