package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	certsv1 "certificate-manager/api/v1"
	"certificate-manager/internal/cert"
)

type CertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	CA cert.CertAuthority
}

//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=certs.k8c.io,resources=certificates/finalizers,verbs=update

func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	var (
		err error

		logger = log.FromContext(ctx)
		crt    = &certsv1.Certificate{}
		result = ctrl.Result{}
	)

	logger.Info("reconciling certificate resources")

	if err = r.Get(ctx, req.NamespacedName, crt); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			logger.Error(err, "unable to fetch certificate resource", "name", req.NamespacedName)
		}

		return result, err
	}

	logger.Info("certificate resource", "name", crt.ObjectMeta.Name)

	return result, nil
}

func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsv1.Certificate{}).
		Complete(r)
}
