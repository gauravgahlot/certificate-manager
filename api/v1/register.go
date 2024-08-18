// Package v1 contains API Schema definitions for the certs v1 API group
// +kubebuilder:object:generate=true
// +groupName=certs.k8c.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion defines the "group" and "version",
	// to uniquely identifiy the API.
	SchemeGroupVersion = schema.GroupVersion{
		Group:   "certs.k8c.io",
		Version: "v1",
	}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
