package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;create;delete;list;update;patch

func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	var (
		err error

		logger  = log.FromContext(ctx)
		crt     = &certsv1.Certificate{}
		result  = ctrl.Result{}
		handler = newRequestHandler(logger, r.Client, r.CA)
	)

	logger.Info("reconciling certificate resources")

	if err = r.Get(ctx, req.NamespacedName, crt); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			logger.Error(err, "unable to fetch certificate resource", "name", req.NamespacedName)
		}

		return result, err
	}

	reconcileAfterDuration, err := handler.updateStatusIfNeeded(ctx, crt)
	if reconcileAfterDuration > 0 {
		logger.Info("reconcile after", "duration", reconcileAfterDuration, "name", req.NamespacedName)

		result.RequeueAfter = reconcileAfterDuration
	}
	if err != nil {
		logger.Error(err, "unable to update certificate status", "name", req.NamespacedName)

		return result, err
	}

	if reconcileAfterDuration == reconcileNone {
		logger.Info("channel reconcile complete", "name", req.NamespacedName)
	}

	return result, nil
}

func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsv1.Certificate{}).
		Owns(&corev1.Secret{}).
		Watches(&corev1.Secret{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
