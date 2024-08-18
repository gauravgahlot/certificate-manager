package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"net"
	"time"
)

const (
	typeRSAKey = "RSA PRIVATE KEY"
	typeCert   = "CERTIFICATE"

	rsaKeySize = 4096
	validDays  = time.Hour * 24 * 365 // 365d
)

type certAuthority struct {
	key          *rsa.PrivateKey
	cert         *x509.Certificate
	countries    []string
	ipAddrs      []net.IP
	validForDays time.Duration
}

func newCertAuthority(cfg *Config) (*certAuthority, error) {
	ca := &certAuthority{
		countries:    []string{"DE", "IN", "US"},
		ipAddrs:      []net.IP{net.ParseIP("127.0.0.1")},
		validForDays: validDays,
	}

	if cfg != nil {
		err := ca.updateCredentials(cfg)
		if err != nil {
			return nil, err
		}
	} else {
		err := ca.newCredentials()
		if err != nil {
			return nil, err
		}
	}

	return ca, nil
}

// GetCredentials returns the base64 encoded CA credentials.
func (ca certAuthority) GetCredentials() ([]byte, []byte) {
	key := x509.MarshalPKCS1PrivateKey(ca.key)
	encodedKey := make([]byte, base64.StdEncoding.EncodedLen(len(key)))
	base64.StdEncoding.Encode(encodedKey, key)

	encodedCert := make([]byte, base64.StdEncoding.EncodedLen(len(ca.cert.Raw)))
	base64.StdEncoding.Encode(encodedCert, ca.cert.Raw)

	return encodedKey, encodedCert
}

// CreateCert creates a new self-signed x509 certificate.
// Returns base64 encoded key and certificate; error otherwise.
func (ca certAuthority) CreateCert(req Request) ([]byte, []byte, error) {
	// generate an RSA root key-pair
	root, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, err
	}

	// create a cert template
	tmpl, err := ca.certTemplate(req, false)
	if err != nil {
		return nil, nil, err
	}

	// generate a self-signed certificate
	cert, err := ca.createCert(tmpl, &root.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	// encode the root key
	encodedKey := encodePem(typeRSAKey, x509.MarshalPKCS1PrivateKey(root))

	// encode the certificate
	encodedCert := encodePem(typeCert, cert)

	return encodedKey, encodedCert, nil
}

// RenewCert renews an existing x509 certificate.
// Returns base64 encoded key and certificate; error otherwise.
func (ca certAuthority) RenewCert(req RenewRequest) ([]byte, []byte, error) {
	return nil, nil, nil
}
