/*
Copyright 2020 The Knative Authors.

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

package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	. "knative.dev/pkg/configmap/testing"
	_ "knative.dev/pkg/system/testing"
)

func TestDefaultsConfigurationFromFile(t *testing.T) {
	cm, example := ConfigMapsFromTestFile(t, DefaultsConfigName)

	if _, err := NewDefaultsConfigFromConfigMap(cm); err != nil {
		t.Fatal("NewDefaultsConfigFromConfigMap(actual) =", err)
	}

	got, err := NewDefaultsConfigFromConfigMap(example)
	if err != nil {
		t.Fatal("NewDefaultsConfigFromConfigMap(example) =", err)
	}
	if want := defaultConfig(); !cmp.Equal(got, want) {
		t.Errorf("Example does not represent default config: diff(-want,+got)\n%s",
			cmp.Diff(want, got))
	}
}

func TestDefaultsConfiguration(t *testing.T) {
	configTests := []struct {
		name         string
		wantErr      bool
		wantDefaults *Defaults
		data         map[string]string
	}{{
		name:         "default configuration",
		wantErr:      false,
		wantDefaults: defaultConfig(),
		data:         map[string]string{},
	}, {
		name:    "specified values",
		wantErr: false,
		wantDefaults: &Defaults{
			RevisionTimeoutSeconds:    123,
			MaxRevisionTimeoutSeconds: 456,
		},
		data: map[string]string{
			"revision-timeout-seconds":     "123",
			"max-revision-timeout-seconds": "456",
		},
	}, {
		name:    "bad revision timeout",
		wantErr: true,
		data: map[string]string{
			"revision-timeout-seconds": "asdf",
		},
	}, {
		name:    "bad max revision timeout",
		wantErr: true,
		data: map[string]string{
			"max-revision-timeout-seconds": "asdf",
		},
	}}

	for _, tt := range configTests {
		t.Run(tt.name, func(t *testing.T) {
			actualDefaults, err := NewDefaultsConfigFromConfigMap(&corev1.ConfigMap{
				Data: tt.data,
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewDefaultsConfigFromConfigMap() error = %v, WantErr %v", err, tt.wantErr)
			}

			if got, want := actualDefaults, tt.wantDefaults; !cmp.Equal(got, want) {
				t.Errorf("Config mismatch: diff(-want,+got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
