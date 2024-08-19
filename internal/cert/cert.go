package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"time"

	"github.com/pkg/errors"
)

const (
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

func newCertAuthority() (*certAuthority, error) {
	ca := &certAuthority{
		countries:    []string{"DE", "IN", "US"},
		ipAddrs:      []net.IP{net.ParseIP("127.0.0.1")},
		validForDays: validDays,
	}

	err := ca.newCredentials()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing CA")
	}

	return ca, nil
}

// CreateCert creates a new self-signed x509 certificate.
// Returns base64 encoded key and certificate; error otherwise.
func (ca certAuthority) CreateCert(req Request) ([]byte, []byte, error) {
	// generate an RSA key-pair
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error generating the private key")
	}

	// encode the private key
	encodedKey, err := encodePKCS1PrivateKey(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error encoding private key")
	}

	// create a cert template
	tmpl, err := ca.certTemplate(req, false)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating a x509 certificate")
	}

	// sign the certificate using the CA
	encodedCert, _, err := ca.signCertificate(tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error signing the x509 certificate")
	}

	return encodedKey, encodedCert, nil
}

// RenewCert renews an existing x509 certificate.
// Returns base64 encoded key and certificate; error otherwise.
func (ca certAuthority) RenewCert(req RenewRequest) ([]byte, []byte, error) {
	return nil, nil, nil
}

// HasCertificateExpired checks whether given base64 encoded
// certificate has expired or not.
func (ca certAuthority) HasCertificateExpired(cert []byte) (bool, error) {
	block, _ := pem.Decode(cert)
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, errors.Wrap(err, "error decoding DER certificate bytes")
	}

	return crt.NotAfter.Before(time.Now()), nil
}
