# knative/networking

This repository contains the Knative Ingress and Certificate CRDs, as well as
their conformance tests. These are our extension points to plugin different
Ingress plugins (Ambassador, Contour, Gloo, Istio, Kong and Kourier), as well as
different ExternalDomainTLS plugins (CertManager and Knative's own HTTP01 challenge
solver).

# Knative Ingress aka KIngress

The Knative Ingress CRD is based on the Kubernetes Ingress CRD, with the
following additions:

1. Traffic splitting
2. Header modification applied to requests based on the traffic split they are
   assigned to.
3. Host header rewrite
4. Traffic redirection predicates based on a regexp based condition on headers.

In addition to these, we previously added Timeout and Retry settings but no
longer used them and may deprecate these parts in future versions.

See
https://knative.dev/docs/install/yaml-install/serving/install-serving-with-yaml/#install-a-networking-layer
for a list of the current supported KIngress implementations.

Check out:

- [pkg/apis/networking/v1alpha1/ingress_types.go](pkg/apis/networking/v1alpha1/ingress_types.go)
  for more information about the KIngress API spec.
- [pkg/apis/networking/v1alpha1/ingress_validation.go](pkg/apis/networking/v1alpha1/ingress_validation.go)
  for more information about the validation logic for KIngress API spec.
- [test/conformance/ingress/README.md](test/conformance/ingress/README.md) for
  the conformance tests and how to run them.

See also:

- http://github.com/knative-extensions/net-contour for the Contour-based
  implementation of KIngress.
- http://github.com/knative-extensions/net-kourier for a dependency-free
  implementation of KIngress based on Envoy proxy.
- http://github.com/knative-extensions/net-istio For the Istio-based implementation
  of KIngress.

# Knative Certificate aka KCert

Knative Certificate CRD is a Knative abstraction for various SSL certificate
provisioning solutions (such as cert-manager or self-signed SSL certificate).

Check out:

- [pkg/apis/networking/v1alpha1/certificate_types.go](pkg/apis/networking/v1alpha1/certificate_types.go)
  for more information about the Certificate API spec.
- [pkg/apis/networking/v1alpha1/certificate_validation.go](pkg/apis/networking/v1alpha1/certificate_validation.go)
  for more information about the validation logic for Certificate API spec.
