package controller

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"sort"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	certsv1 "certificate-manager/api/v1"
	"certificate-manager/internal/cert"
)

func (rh *requestHandler) getSecret(ctx context.Context, key types.NamespacedName, obj *corev1.Secret) error {
	err := rh.client.Get(ctx, key, obj)
	if err != nil {
		return err
	}

	return nil
}

func (rh *requestHandler) createSecret(ctx context.Context, obj *certsv1.Certificate) error {
	key, crt, err := rh.ca.IssueCert(cert.Request{
		Organization: obj.Spec.Organization,
		DNSName:      obj.Spec.DNSName,
		ValidForDays: obj.Spec.ValidForDays,
		AltNames:     obj.Spec.AltNames,
	})
	if err != nil {
		return err
	}

	sec := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:        obj.Spec.SecretRef.Name,
			Namespace:   obj.ObjectMeta.Namespace,
			Labels:      obj.Labels,
			Annotations: obj.Annotations,
		},
		Immutable: &isImmutable,
		Type:      corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSPrivateKeyKey: key,
			corev1.TLSCertKey:       crt,
		},
	}

	if err := controllerutil.SetControllerReference(obj, sec, rh.client.Scheme()); err != nil {
		return err
	}

	return rh.client.Create(ctx, sec)
}

func (rh *requestHandler) hasCertificateExpired(sec corev1.Secret) (bool, error) {
	if data, ok := sec.Data[tlsCert]; ok {
		return rh.ca.HasCertificateExpired(data)
	}

	return false, nil
}

func certificateHasChanges(n *certsv1.Certificate, o *certsv1.Certificate) bool {
	sort.Strings(n.Spec.AltNames)
	sort.Strings(o.Spec.AltNames)

	for i, name := range n.Spec.AltNames {
		if name != o.Spec.AltNames[i] {
			return true
		}
	}

	return n.Spec.DNSName != o.Spec.DNSName ||
		n.Spec.Organization != o.Spec.Organization ||
		n.Spec.ValidForDays != o.Spec.ValidForDays ||
		n.Spec.SecretRef.Name != o.Spec.SecretRef.Name
}

func getCertFromExternalWorld(obj *corev1.Secret, cert *certsv1.Certificate) error {
	crt, err := getX509Certificate(obj.Data[tlsCert])
	if err != nil {
		return err
	}

	cert.Spec.DNSName = crt.Subject.CommonName
	cert.Spec.Organization = crt.Subject.Organization[0]
	cert.Spec.AltNames = crt.DNSNames
	cert.Spec.ValidForDays = getValidForDays(crt.NotAfter)
	cert.Spec.SecretRef.Name = obj.ObjectMeta.Name

	return nil
}

func getX509Certificate(crtBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(crtBytes)
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding DER certificate bytes")
	}

	return crt, nil
}

func getValidForDays(notAfter time.Time) int {
	current := time.Now().Truncate(24 * time.Hour)
	notAfter = notAfter.Truncate(24 * time.Hour)

	if notAfter.After(current) {
		duration := notAfter.Sub(current)
		return int(duration.Hours() / 24)
	}

	return -1
}
