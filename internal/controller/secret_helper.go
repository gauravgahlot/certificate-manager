package controller

import (
	"context"

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
	key, crt, err := rh.ca.CreateCert(cert.Request{
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
			Name:      obj.Spec.SecretRef.Name,
			Namespace: obj.ObjectMeta.Namespace,
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
