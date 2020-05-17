package networking

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"knative.dev/pkg/apis"
)

func ValidateNamespacedObjectReference(p *corev1.ObjectReference) *apis.FieldError {
	if p == nil {
		return nil
	}
	errs := apis.CheckDisallowedFields(*p, *NamespacedObjectReferenceMask(p))

	if p.APIVersion == "" {
		errs = errs.Also(apis.ErrMissingField("apiVersion"))
	} else if verrs := validation.IsQualifiedName(p.APIVersion); len(verrs) != 0 {
		errs = errs.Also(apis.ErrInvalidValue(strings.Join(verrs, ", "), "apiVersion"))
	}
	if p.Kind == "" {
		errs = errs.Also(apis.ErrMissingField("kind"))
	} else if verrs := validation.IsCIdentifier(p.Kind); len(verrs) != 0 {
		errs = errs.Also(apis.ErrInvalidValue(strings.Join(verrs, ", "), "kind"))
	}
	if p.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	} else if verrs := validation.IsDNS1123Label(p.Name); len(verrs) != 0 {
		errs = errs.Also(apis.ErrInvalidValue(strings.Join(verrs, ", "), "name"))
	}
	return errs
}
