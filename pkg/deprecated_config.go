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

package pkg

import "knative.dev/networking/pkg/config"

type (
	// DomainTemplateValues are the available properties people can choose from
	// in their Route's "DomainTemplate" golang template sting.
	// We could add more over time - e.g. RevisionName if we thought that
	// might be of interest to people.
	// Deprecated: use config.DomainTemplateValues
	DomainTemplateValues = config.DomainTemplateValues

	// TagTemplateValues are the available properties people can choose from
	// in their Route's "TagTemplate" golang template sting.
	// Deprecated: use config.TagTemplateValues
	TagTemplateValues = config.TagTemplateValues

	// Config contains the networking configuration defined in the
	// network config map.
	// Deprecated: use config.Config
	Config = config.Config

	// HTTPProtocol indicates a type of HTTP endpoint behavior
	// that Knative ingress could take.
	// Deprecated: config.HTTPProtocol
	HTTPProtocol = config.HTTPProtocol
)

var (
	// NewConfigFromConfigMap creates a Config from the supplied ConfigMap
	// Deprecated: use config.NewFromConfigMap
	NewConfigFromConfigMap = config.NewFromConfigMap

	// NewConfigFromMap creates a Config from the supplied data.
	// Deprecated: use config.NewFromMap
	NewConfigFromMap = config.NewFromMap

	// HTTPEnabled represents HTTP protocol is enabled in Knative ingress.
	// Deprecated: use config.HTTPEnabled
	HTTPEnabled = config.HTTPEnabled

	// HTTPDisabled represents HTTP protocol is disabled in Knative ingress.
	// Deprecated: use config.HTTPDisabled
	HTTPDisabled = config.HTTPDisabled

	// HTTPRedirected represents HTTP connection is redirected to HTTPS in Knative ingress.
	// Deprecated: use config.HTTPRedirected
	HTTPRedirected = config.HTTPRedirected

	// ConfigName is the name of the configmap containing all
	// customizations for networking features.
	// Deprecated: use config.ConfigName
	ConfigName = config.ConfigName

	// DefaultIngressClassKey is the name of the configuration entry
	// that specifies the default Ingress.
	// Deprecated: use config.DefaultIngressClassKey
	DefaultIngressClassKey = config.DefaultIngressClassKey

	// DefaultCertificateClassKey is the name of the configuration entry
	// that specifies the default Certificate.
	// Deprecated: use config.DefaultCertificateClassKey
	DefaultCertificateClassKey = config.DefaultCertificateClassKey

	// IstioIngressClassName value for specifying knative's Istio
	// Ingress reconciler.
	// Deprecated: use config.IstioIngressClassName
	IstioIngressClassName = config.IstioIngressClassName

	// CertManagerCertificateClassName value for specifying Knative's Cert-Manager
	// Certificate reconciler.
	// Deprecated: use config.IstioIngressClassName
	CertManagerCertificateClassName = config.IstioIngressClassName

	// TagHeaderBasedRoutingKey is the name of the configuration entry
	// that specifies enabling tag header based routing or not.
	// Deprecated: use config.TagHeaderBasedRoutingKey
	TagHeaderBasedRoutingKey = config.TagHeaderBasedRoutingKey

	// DomainTemplateKey is the name of the configuration entry that
	// specifies the golang template string to use to construct the
	// Knative service's DNS name.
	// Deprecated: use config.DomainTemplateKey
	DomainTemplateKey = config.DomainTemplateKey

	// TagTemplateKey is the name of the configuration entry that
	// specifies the golang template string to use to construct the
	// hostname for a Route's tag.
	// Deprecated: use config.TagTemplateKey
	TagTemplateKey = config.TagTemplateKey

	// RolloutDurationKey is the name of the configuration entry
	// that specifies the default duration of the configuration rollout.
	// Deprecated: use config.RolloutDurationKey
	RolloutDurationKey = config.RolloutDurationKey

	// DefaultDomainTemplate is the default golang template to use when
	// constructing the Knative Route's Domain(host)
	// Deprecated: use config.DefaultDomainTemplate
	DefaultDomainTemplate = config.DefaultDomainTemplate

	// DefaultTagTemplate is the default golang template to use when
	// constructing the Knative Route's tag names.
	// Deprecated: use config.DefaultTagTemplate
	DefaultTagTemplate = config.DefaultTagTemplate

	// AutocreateClusterDomainClaimsKey is the key for the
	// AutocreateClusterDomainClaims property.
	// Deprecated: use config.AutocreateClusterDomainClaimsKey
	AutocreateClusterDomainClaimsKey = config.AutocreateClusterDomainClaimsKey

	// AutoTLSKey is the name of the configuration entry
	// that specifies enabling auto-TLS or not.
	// Deprecated: use config.AutoTLSKey
	AutoTLSKey = config.AutoTLSKey

	// HTTPProtocolKey is the name of the configuration entry that
	// specifies the HTTP endpoint behavior of Knative ingress.
	// Deprecated: use config.HTTPProtocolKey
	HTTPProtocolKey = config.HTTPProtocolKey

	// EnableMeshPodAddressabilityKey is the config for enabling pod addressability in mesh.
	// Deprecated: use config.EnableMeshPodAddressabilityKey
	EnableMeshPodAddressabilityKey = config.EnableMeshPodAddressabilityKey

	// DefaultExternalSchemeKey is the config for defining the scheme of external URLs.
	// Deprecated: use config.DefaultExternalSchemeKey
	DefaultExternalSchemeKey = config.DefaultExternalSchemeKey
)
