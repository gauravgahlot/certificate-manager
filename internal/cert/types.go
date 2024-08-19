package cert

// CertAuthority defines a certificate authority.
type CertAuthority interface {
	// CreateCert creates a new self-signed x509 certificate.
	// Returns base64 encoded key and certificate; error otherwise.
	CreateCert(Request) (key []byte, crt []byte, e error)

	// RenewCert renews an existing x509 certificate.
	// Returns base64 encoded key and certificate; error otherwise.
	RenewCert(RenewRequest) (key []byte, crt []byte, e error)

	// HasCertificateExpired checks whether given base64 encoded
	// certificate has expired or not.
	HasCertificateExpired([]byte) (bool, error)
}

// Request holds the required fields for generating a certificate.
type Request struct {
	ValidForDays int
	Organization string
	DNSName      string
	AltNames     []string
}

type RenewRequest struct {
	Request

	// TLSKey is the base64 encoded key.
	TLSKey []byte

	// TLSCert is the base64 encoded certificate.
	TLSCert []byte
}

// Authority initializes and returns a Certificate Authority.
func Authority() (CertAuthority, error) {
	return newCertAuthority()
}
