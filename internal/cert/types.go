package cert

// CertAuthority defines a certificate authority.
type CertAuthority interface {
	// GetCredentials returns the base64 encoded CA credentials.
	GetCredentials() (key []byte, crt []byte)

	// CreateCert creates a new self-signed x509 certificate.
	// Returns base64 encoded key and certificate; error otherwise.
	CreateCert(Request) (key []byte, crt []byte, e error)

	// RenewCert renews an existing x509 certificate.
	// Returns base64 encoded key and certificate; error otherwise.
	RenewCert(RenewRequest) (key []byte, crt []byte, e error)
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

// Config holds the existing CA credentials.
type Config struct {
	// CAKey is the base64 encoded key.
	CAKey []byte

	// CACert is the base64 encoded certificate.
	CACert []byte
}

// Authority initializes and returns a Certificate Authority.
func Authority(cfg *Config) (CertAuthority, error) {
	return newCertAuthority(cfg)
}
