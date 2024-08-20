package cert

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

const (
	shiftBits = 128

	typeRSAKey = "RSA PRIVATE KEY"
	typeCert   = "CERTIFICATE"
)

// newCredentials creates new CA credentials.
func (ca *certAuthority) newCredentials() error {
	// generate an RSA root key-pair for CA
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return errors.Wrap(err, "error generating the private key")
	}

	tmpl, err := ca.certTemplate(Request{Organization: "certificate-manager"}, true)
	if err != nil {
		return errors.Wrap(err, "error creating a x509 certificate")
	}

	// self sign root CA
	_, cert, err := ca.signCertificate(tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return errors.Wrap(err, "error signing the x509 certificate")
	}

	ca.key = key
	ca.cert = cert

	return nil
}

// certTemplate returns a x509 Certificate with required fields.
func (ca certAuthority) certTemplate(req Request, isCA bool) (*x509.Certificate, error) {
	snLimit := new(big.Int).Lsh(big.NewInt(1), shiftBits)
	sn, err := rand.Int(rand.Reader, snLimit)
	if err != nil {
		return nil, errors.Wrap(err, "error generating a serial number")
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	if isCA {
		keyUsage |= x509.KeyUsageCertSign
	}

	tmpl := &x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			CommonName:   req.DNSName,
			Country:      ca.countries,
			Organization: []string{req.Organization},
		},
		NotBefore:             time.Now(),
		KeyUsage:              keyUsage,
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		DNSNames:              req.AltNames,
	}

	if req.ValidForDays <= 0 || isCA {
		tmpl.NotAfter = tmpl.NotBefore.Add(ca.validForDays)
	} else {
		tmpl.NotAfter = tmpl.NotBefore.Add(time.Hour * 24 * time.Duration(req.ValidForDays))
	}

	if !isCA {
		tmpl.IPAddresses = ca.ipAddrs
	}

	return tmpl, nil
}

// generate a self-signed certificate
func (ca certAuthority) signCertificate(tmpl *x509.Certificate,
	issuerCert *x509.Certificate,
	pubKey crypto.PublicKey,
	signerKey interface{}) ([]byte, *x509.Certificate, error) {

	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, issuerCert, pubKey, signerKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating x509 certificate")
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error decoding DER certificate bytes")
	}

	encoded, err := encodeX509(cert)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error encoding certificate PEM")
	}

	return encoded, cert, nil
}

// encodeX509 will encode a single *x509.Certificate into PEM format.
func encodeX509(cert *x509.Certificate) ([]byte, error) {
	pemBytes := bytes.NewBuffer([]byte{})

	if err := pem.Encode(pemBytes, &pem.Block{
		Type:  typeCert,
		Bytes: cert.Raw,
	}); err != nil {
		return nil, err
	}

	return pemBytes.Bytes(), nil
}

// encodePKCS1PrivateKey will marshal a RSA private key into x509 PEM format.
func encodePKCS1PrivateKey(pk *rsa.PrivateKey) ([]byte, error) {
	pemBytes := bytes.NewBuffer([]byte{})

	if err := pem.Encode(pemBytes, &pem.Block{
		Type:  typeRSAKey,
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}); err != nil {
		return nil, err
	}

	return pemBytes.Bytes(), nil
}
