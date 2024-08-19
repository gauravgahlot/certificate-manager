package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certsv1 "certificate-manager/api/v1"
	"certificate-manager/internal/cert"
)

type requestHandler struct {
	logger logr.Logger
	client client.Client
	ca     cert.CertAuthority
}

func newRequestHandler(logger logr.Logger,
	client client.Client, ca cert.CertAuthority) *requestHandler {

	return &requestHandler{logger, client, ca}
}

func (rh requestHandler) updateStatusIfNeeded(
	ctx context.Context, cert *certsv1.Certificate) (time.Duration, error) {

	// if it's a new certificate or the credentials have expired
	// then create a new secret with valid credentials
	if cert.Status.State == "" || cert.Status.State == certsv1.StateExpired {
		if err := rh.createSecret(ctx, cert); err != nil {
			return reconcileInAMinute, err
		}

		cert.Status.State = certsv1.StateValid
		if err := rh.client.Status().Update(ctx, cert); err != nil {
			return reconcileShortly, err
		}

		return reconcileNone, nil
	}

	// at this point we have a secret which may or maynot have valid credentials
	key := client.ObjectKey{
		Namespace: cert.Namespace,
		Name:      cert.Spec.SecretRef.Name,
	}

	var sec corev1.Secret
	err := rh.getSecret(ctx, key, &sec)
	if err != nil {
		// if the secret is not found, set the state as Expired
		// so that a new one can be created
		if errors.IsNotFound(err) {
			cert.Status.State = certsv1.StateExpired

			return reconcileShortly, rh.client.Status().Update(ctx, cert)
		}

		rh.logger.Error(err, "unable to fetch secret", "name", key.String())

		return reconcileShortly, err
	}

	// check if the certificate has expired
	expired, err := rh.hasCertificateExpired(sec)
	if err != nil {
		rh.logger.Error(err, "unable to validate secret credentials", "name", key.String())

		return reconcileShortly, err
	}

	if expired {
		cert.Status.State = certsv1.StateExpired
		rh.logger.Info("secret credentials have expired", "name", key.String())

		return reconcileShortly, rh.client.Status().Update(ctx, cert)
	}

	return reconcileNone, nil
}
