package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateSpec defines the desired state of the Certificate.
type CertificateSpec struct {
	// Name of the organization.
	Organization string `json:"organization"`

	// The DNS name for which the certificate should be issued.
	DNSName string `json:"dnsName"`

	// The number of days until the certificate expires.
	// +kubebuilder:validation:Minimum=7
	// +kubebuilder:default=365
	ValidForDays int `json:"validForDays,omitempty"`

	// The number of days the certificate should be renewed before it expires.
	// +kubebuilder:validation:Minimum=7
	// +kubebuilder:default:15
	RenewBeforeDays int `json:"renewBeforeDays,omitempty"`

	// Subject alternate names, other than DNSName.
	AltNames []string `json:"altNames,omitempty"`

	// A reference to the Secret object in which the certificate is stored.
	SecretRef SecretRef `json:"secretRef"`
}

type SecretRef struct {
	Name string `json:"name"`
}

// CertificateStatus defines the observed state of the certificate.
type CertificateStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Certificate is the schema for the certs API.
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertificateList contains a list of Certificate.
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Certificate `json:"items"`
}

// Register types to the scheme builder, so they can be added to the scheme.
func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
