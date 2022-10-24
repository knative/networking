/*
Copyright 2022 The Knative Authors

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
	"bytes"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/lru"

	. "knative.dev/pkg/configmap/testing"
)

func TestOurConfig(t *testing.T) {
	cm, example := ConfigMapsFromTestFile(t, ConfigMapName)

	if _, err := NewConfigFromMap(cm.Data); err != nil {
		t.Error("NewConfigFromConfigMap(actual) =", err)
	}
	if got, err := NewConfigFromMap(example.Data); err != nil {
		t.Error("NewConfigFromConfigMap(example) =", err)
	} else if want := defaultConfig(); !cmp.Equal(got, want) {
		t.Errorf("ExampleConfig does not match default config: (-want,+got):\n%s", cmp.Diff(want, got))
	}
}

func TestConfiguration(t *testing.T) {
	const nonDefaultDomainTemplate = "{{.Namespace}}.{{.Name}}.{{.Domain}}"
	ignoreDT := cmpopts.IgnoreFields(Config{}, "DomainTemplate")

	networkConfigTests := []struct {
		name       string
		wantErr    bool
		wantConfig *Config
		data       map[string]string
	}{{
		name:       "network configuration with no network input",
		wantErr:    false,
		wantConfig: defaultConfig(),
	}, {
		name: "network configuration with non-default ingress type",
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.DefaultIngressClass = "foo-ingress"
			return c
		}(),
	}, {
		name: "network configuration with proxy protocol probe enabled",
		data: map[string]string{
			ProxyProtocolProbeEnabled: "true",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.ProxyProtocolProbeEnabled = true
			return c
		}(),
	}, {
		name: "network configuration with proxy protocol filter",
		data: map[string]string{
			ProxyProtocolFilter: "fizz",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.ProxyProtocolFilter = "fizz"
			return c
		}(),
	}, {
		name: "network configuration with non-default rollout duration",
		data: map[string]string{
			RolloutDurationKey: "211",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.RolloutDurationSecs = 211
			return c
		}(),
	}, {
		name: "network configuration with bad rollout duration",
		data: map[string]string{
			RolloutDurationKey: "mil novecientos ochenta y dos",
		},
		wantErr: true,
	}, {
		name: "network configuration with negative rollout duration",
		data: map[string]string{
			RolloutDurationKey: "-444",
		},
		wantErr: true,
	}, {
		name: "network configuration with non-default autocreateClusterDomainClaim value",
		data: map[string]string{
			AutocreateClusterDomainClaimsKey: "false",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.AutocreateClusterDomainClaims = false
			return c
		}(),
	}, {
		name: "network configuration with invalid autocreateClusterDomainClaim value",
		data: map[string]string{
			AutocreateClusterDomainClaimsKey: "salad",
		},
		wantErr: true,
	}, {
		name: "network configuration with non-Cert-Manager Certificate type",
		data: map[string]string{
			DefaultCertificateClassKey: "foo-cert",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.DefaultCertificateClass = "foo-cert"
			return c
		}(),
	}, {
		name: "network configuration with configured wildcard cert label selector",
		data: map[string]string{
			NamespaceWildcardCertSelectorKey: "matchExpressions:\n- key: networking.knative.dev/disableWildcardCert\n  operator: NotIn\n  values: [\"true\"]",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.NamespaceWildcardCertSelector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "networking.knative.dev/disableWildcardCert",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"true"},
				}},
			}
			return c
		}(),
	}, {
		name: "network configuration with diff domain template",
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      nonDefaultDomainTemplate,
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.DefaultIngressClass = "foo-ingress"
			c.DomainTemplate = nonDefaultDomainTemplate
			return c
		}(),
	}, {
		name:    "network configuration with blank domain template",
		wantErr: true,
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      "",
		},
	}, {
		name:    "network configuration with bad domain template",
		wantErr: true,
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			// This is missing a closing brace.
			DomainTemplateKey: "{{.Namespace}.{{.Name}}.{{.Domain}}",
		},
	}, {
		name:    "network configuration with bad domain template",
		wantErr: true,
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			// This is missing a closing brace.
			DomainTemplateKey: "{{.Namespace}.{{.Name}}.{{.Domain}}",
		},
	}, {
		name:    "network configuration with bad url",
		wantErr: true,
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			// Paths are disallowed
			DomainTemplateKey: "{{.Domain}}/{{.Namespace}}/{{.Name}}.",
		},
	}, {
		name:    "network configuration with bad variable",
		wantErr: true,
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			// Bad variable
			DomainTemplateKey: "{{.Name}}.{{.NAmespace}}.{{.Domain}}",
		},
	}, {
		name: "network configuration with Auto TLS enabled",
		data: map[string]string{
			AutoTLSKey: "enabled",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.AutoTLS = true
			return c
		}(),
	}, {
		name: "network configuration with Auto TLS disabled",
		data: map[string]string{
			AutoTLSKey: "disabled",
		},
		wantErr:    false,
		wantConfig: defaultConfig(),
	}, {
		name: "network configuration with HTTPProtocol disabled",
		data: map[string]string{
			AutoTLSKey:      "enabled",
			HTTPProtocolKey: "Disabled",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.AutoTLS = true
			c.HTTPProtocol = HTTPDisabled
			return c
		}(),
	}, {
		name: "network configuration with HTTPProtocol redirected",
		data: map[string]string{
			AutoTLSKey:      "enabled",
			HTTPProtocolKey: "Redirected",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.AutoTLS = true
			c.HTTPProtocol = HTTPRedirected
			return c
		}(),
	}, {
		name: "network configuration with HTTPProtocol bad",
		data: map[string]string{
			AutoTLSKey:      "enabled",
			HTTPProtocolKey: "under-the-bridge",
		},
		wantErr: true,
	}, {
		name: "network configuration with enabled pod-addressability",
		data: map[string]string{
			EnableMeshPodAddressabilityKey: "true",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.EnableMeshPodAddressability = true
			return c
		}(),
	}, {
		name: "network configuration with enabled mesh compatibility mode",
		data: map[string]string{
			MeshCompatibilityModeKey: "enabled",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.MeshCompatibilityMode = MeshCompatibilityModeEnabled
			return c
		}(),
	}, {
		name: "network configuration with disabled mesh compatibility mode",
		data: map[string]string{
			MeshCompatibilityModeKey: "disabled",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.MeshCompatibilityMode = MeshCompatibilityModeDisabled
			return c
		}(),
	}, {
		name: "network configuration with overridden external and internal scheme",
		data: map[string]string{
			DefaultExternalSchemeKey: "https",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.DefaultExternalScheme = "https"
			return c
		}(),
	}, {
		name: "network configuration with activator-ca and activator-san",
		data: map[string]string{
			InternalEncryptionKey: "true",
		},
		wantErr: false,
		wantConfig: func() *Config {
			c := defaultConfig()
			c.InternalEncryption = true
			return c
		}(),
	}, {
		name: "legacy keys",
		data: map[string]string{
			"ingress.class":         "1",
			"certificate.class":     "2",
			"domainTemplate":        "3",
			"tagTemplate":           "4",
			"rolloutDuration":       "5",
			"defaultExternalScheme": "6",

			"autocreateClusterDomainClaims": "true",
			"httpProtocol":                  "redirected",
			"autoTLS":                       "enabled",
		},
		wantConfig: &Config{
			DefaultIngressClass:     "1",
			DefaultCertificateClass: "2",
			DomainTemplate:          "3",
			TagTemplate:             "4",
			RolloutDurationSecs:     5,
			DefaultExternalScheme:   "6",

			AutocreateClusterDomainClaims: true,
			HTTPProtocol:                  HTTPRedirected,
			AutoTLS:                       true,

			// This is defaulted
			MeshCompatibilityMode: MeshCompatibilityModeAuto,
		},
	}, {
		name: "newer keys take precedence over legacy keys",
		data: map[string]string{
			"ingress.class":         "1",
			"certificate.class":     "2",
			"domainTemplate":        "3",
			"tagTemplate":           "4",
			"rolloutDuration":       "5",
			"defaultExternalScheme": "6",

			"autocreateClusterDomainClaims": "true",
			"httpProtocol":                  "redirected",
			"autoTLS":                       "enabled",

			DefaultIngressClassKey:     "7",
			DefaultCertificateClassKey: "8",
			DomainTemplateKey:          "9",
			TagTemplateKey:             "10",
			RolloutDurationKey:         "11",
			DefaultExternalSchemeKey:   "12",

			AutocreateClusterDomainClaimsKey: "false",
			HTTPProtocolKey:                  "enabled",
			AutoTLSKey:                       "disabled",
		},
		wantConfig: &Config{
			DefaultIngressClass:     "7",
			DefaultCertificateClass: "8",
			DomainTemplate:          "9",
			TagTemplate:             "10",
			RolloutDurationSecs:     11,
			DefaultExternalScheme:   "12",

			AutocreateClusterDomainClaims: false,
			HTTPProtocol:                  HTTPEnabled,
			AutoTLS:                       false,

			// This is defaulted
			MeshCompatibilityMode: MeshCompatibilityModeAuto,
		},
	}}

	for _, tt := range networkConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			actualConfig, err := NewConfigFromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			data := DomainTemplateValues{
				Name:      "foo",
				Namespace: "bar",
				Domain:    "baz.com",
			}
			want := mustExecute(t, tt.wantConfig.GetDomainTemplate(), data)
			got := mustExecute(t, actualConfig.GetDomainTemplate(), data)
			if got != want {
				t.Errorf("DomainTemplate(data) = %s, wanted %s", got, want)
			}

			if diff := cmp.Diff(actualConfig, tt.wantConfig, ignoreDT); diff != "" {
				t.Fatalf("diff (-want,+got) %v", diff)
			}
		})
	}
}

func TestTemplateCaching(t *testing.T) {
	// Reset the template cache, to ensure size change.
	templateCache = lru.New(10)

	const anotherTemplate = "{{.Namespace}}.{{.Name}}.{{.Domain}}.sad"
	actualConfig, err := NewConfigFromMap(map[string]string{
		DomainTemplateKey: anotherTemplate,
	})
	if err != nil {
		t.Fatal("Config parsing failure =", err)
	}
	if got, want := actualConfig.DomainTemplate, anotherTemplate; got != want {
		t.Errorf("DomainTemplate = %q, want: %q", got, want)
	}
	if got, want := templateCache.Len(), 2; got != want {
		t.Errorf("Cache size = %d, want = %d", got, want)
	}

	// Reset to default. And make sure it is cached.
	actualConfig, err = NewConfigFromMap(map[string]string{})
	if err != nil {
		t.Fatal("Config parsing failure =", err)
	}

	if got, want := actualConfig.DomainTemplate, DefaultDomainTemplate; got != want {
		t.Errorf("DomainTemplate = %q, want: %q", got, want)
	}
	if got, want := templateCache.Len(), 3; got != want {
		t.Errorf("Cache size = %d, want = %d", got, want)
	}
}

func TestAnnotationsInDomainTemplate(t *testing.T) {
	networkConfigTests := []struct {
		name               string
		wantErr            bool
		wantDomainTemplate string
		config             map[string]string
		data               DomainTemplateValues
	}{{
		name:               "network configuration with annotations in template",
		wantErr:            false,
		wantDomainTemplate: "foo.sub1.baz.com",
		config: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      `{{.Name}}.{{ index .Annotations "sub"}}.{{.Domain}}`,
		},
		data: DomainTemplateValues{
			Name:      "foo",
			Namespace: "bar",
			Annotations: map[string]string{
				"sub": "sub1"},
			Domain: "baz.com"},
	}, {
		name:               "network configuration without annotations in template",
		wantErr:            false,
		wantDomainTemplate: "foo.bar.baz.com",
		config: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      `{{.Name}}.{{.Namespace}}.{{.Domain}}`,
		},
		data: DomainTemplateValues{
			Name:      "foo",
			Namespace: "bar",
			Domain:    "baz.com"},
	}}

	for _, tt := range networkConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			actualConfig, err := NewConfigFromMap(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got := mustExecute(t, actualConfig.GetDomainTemplate(), tt.data)
			if got != tt.wantDomainTemplate {
				t.Errorf("DomainTemplate(data) = %s, wanted %s", got, tt.wantDomainTemplate)
			}
		})
	}
}

func TestLabelsInDomainTemplate(t *testing.T) {
	networkConfigTests := []struct {
		name               string
		data               map[string]string
		templateValue      DomainTemplateValues
		wantErr            bool
		wantDomainTemplate string
	}{{
		name:               "network configuration with labels in template",
		wantErr:            false,
		wantDomainTemplate: "foo.sub1.baz.com",
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      `{{.Name}}.{{ index .Labels "sub"}}.{{.Domain}}`,
		},
		templateValue: DomainTemplateValues{
			Name:      "foo",
			Namespace: "bar",
			Labels: map[string]string{
				"sub": "sub1"},
			Domain: "baz.com"},
	}, {
		name:               "network configuration without labels in template",
		wantErr:            false,
		wantDomainTemplate: "foo.bar.baz.com",
		data: map[string]string{
			DefaultIngressClassKey: "foo-ingress",
			DomainTemplateKey:      `{{.Name}}.{{.Namespace}}.{{.Domain}}`,
		},
		templateValue: DomainTemplateValues{
			Name:      "foo",
			Namespace: "bar",
			Domain:    "baz.com"},
	}}

	for _, tt := range networkConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			actualConfig, err := NewConfigFromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromConfigMap() error = %v, WantErr? %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got := mustExecute(t, actualConfig.GetDomainTemplate(), tt.templateValue)
			if got != tt.wantDomainTemplate {
				t.Errorf("DomainTemplate(data) = %s, want: %s", got, tt.wantDomainTemplate)
			}
		})
	}
}

func mustExecute(t *testing.T, tmpl *template.Template, data interface{}) string {
	t.Helper()
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Error("Error executing the DomainTemplate:", err)
	}
	return buf.String()
}
