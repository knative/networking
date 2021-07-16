/*
Copyright 2021 The Knative Authors

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
package config

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	lru "github.com/hashicorp/golang-lru"

	. "knative.dev/pkg/configmap/testing"
	_ "knative.dev/pkg/system/testing"
)

func TestDefaultTemplateCompile(t *testing.T) {
	// Verify the default templates are valid.
	if _, err := template.New("domain-template").Parse(DefaultDomainTemplate); err != nil {
		t.Error("DefaultDomainTemplate did not compile: ", err)
	}
	if _, err := template.New("tag-template").Parse(DefaultTagTemplate); err != nil {
		t.Error("DefaultTagTemplate did not compile: ", err)
	}
}

func TestOurConfig(t *testing.T) {
	cm, example := ConfigMapsFromTestFile(t, ConfigName)

	if _, err := NewFromMap(cm.Data); err != nil {
		t.Error("NewConfigFromConfigMap(actual) =", err)
	}
	if got, err := NewFromMap(example.Data); err != nil {
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
		name: "network configuration with overridden external and internal scheme",
		data: map[string]string{
			DefaultExternalSchemeKey: "https",
		},
		wantConfig: func() *Config {
			c := defaultConfig()
			c.DefaultExternalScheme = "https"
			return c
		}(),
	}}

	for _, tt := range networkConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			actualConfigCM, err := NewFromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}

			actualConfig, err := NewFromMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}
			if diff := cmp.Diff(actualConfigCM, actualConfig); diff != "" {
				t.Errorf("Config mismatch: diff(-want,+got):\n%s", diff)
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
				t.Fatalf("want %v, but got %v", tt.wantConfig, actualConfig)
			}
		})
	}
}

func TestTemplateCaching(t *testing.T) {
	// Reset the template cache, to ensure size change.
	templateCache, _ = lru.New(10)

	const anotherTemplate = "{{.Namespace}}.{{.Name}}.{{.Domain}}.sad"
	actualConfig, err := NewFromMap(map[string]string{
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
	actualConfig, err = NewFromMap(map[string]string{})
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
			actualConfigCM, err := NewFromMap(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}

			actualConfig, err := NewFromMap(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewConfigFromMap() error = %v, WantErr %v",
					err, tt.wantErr)
			}

			if diff := cmp.Diff(actualConfigCM, actualConfig); diff != "" {
				t.Errorf("Config mismatch: diff(-want,+got):\n%s", diff)
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
			actualConfig, err := NewFromMap(tt.data)
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
