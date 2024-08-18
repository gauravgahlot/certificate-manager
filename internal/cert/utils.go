package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"time"
)

const shiftBits = 128

// updateCredentials loads the existing CA credentials.
func (ca *certAuthority) updateCredentials(cfg *Config) error {
	decodedKey := make([]byte, base64.StdEncoding.EncodedLen(len(cfg.CAKey)))
	_, err := base64.StdEncoding.Decode(decodedKey, cfg.CAKey)
	if err != nil {
		return err
	}

	caKey, err := x509.ParsePKCS1PrivateKey(decodedKey)
	if err != nil {
		return err
	}
	ca.key = caKey

	decodedCert := make([]byte, base64.StdEncoding.EncodedLen(len(cfg.CACert)))
	_, err = base64.StdEncoding.Decode(decodedCert, cfg.CACert)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(decodedCert)
	if err != nil {
		return err
	}
	ca.cert = cert

	return nil
}

// newCredentials creates new CA credentials.
func (ca *certAuthority) newCredentials() error {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return err
	}

	tmpl, err := ca.certTemplate(Request{Organization: "certificate-manager"}, true)
	if err != nil {
		return err
	}

	ca.key = key
	ca.cert = tmpl

	certBytes, err := ca.createCert(tmpl, &key.PublicKey)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return err
	}
	ca.cert = cert

	return nil
}

// certTemplate returns a x509 Certificate with required fields.
func (ca certAuthority) certTemplate(req Request, isCA bool) (*x509.Certificate, error) {
	snLimit := new(big.Int).Lsh(big.NewInt(1), shiftBits)
	sn, err := rand.Int(rand.Reader, snLimit)
	if err != nil {
		return nil, err
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
		DNSNames:              append(req.AltNames, req.DNSName),
	}

	if req.ValidForDays <= 0 || isCA {
		tmpl.NotAfter = time.Now().Add(ca.validForDays)
	} else {
		tmpl.NotAfter = time.Now().Add(time.Hour * 24 * time.Duration(req.ValidForDays))
	}

	if !isCA {
		tmpl.IPAddresses = ca.ipAddrs
	}

	return tmpl, nil
}

// generate a self-signed certificate
func (ca certAuthority) createCert(tmpl *x509.Certificate, pubKey any) ([]byte, error) {
	cert, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, pubKey, ca.key)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// encodePem returns the base64 encoded permission block
func encodePem(pemType string, content []byte) []byte {
	block := pem.EncodeToMemory(&pem.Block{
		Type:  pemType,
		Bytes: content,
	})

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(block)))
	base64.StdEncoding.Encode(encoded, block)

	return encoded
}
