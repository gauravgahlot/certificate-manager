package cert

// CertAuthority defines a certificate authority.
type CertAuthority interface {
	// IssueCert issues a self-signed x509 certificate.
	// Returns base64 encoded key and certificate; error otherwise.
	IssueCert(Request) (key []byte, crt []byte, e error)

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

// Authority initializes and returns a Certificate Authority.
func Authority() (CertAuthority, error) {
	return newCertAuthority()
}
