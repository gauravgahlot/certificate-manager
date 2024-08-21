//go:build e2e

package controller_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"go.uber.org/mock/gomock"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	certsv1 "certificate-manager/api/v1"
	"certificate-manager/internal/cert"
	"certificate-manager/internal/cert/mocks"
	"certificate-manager/internal/controller"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	mockCtrl  *gomock.Controller
	ca        *mocks.MockCertAuthority
	tlsKey    []byte
	tlsCrt    []byte
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.29.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	// generate a certificate for e2e test
	trueCA, _ := cert.Authority()
	tlsKey, tlsCrt, _ = trueCA.IssueCert(cert.Request{
		DNSName:      "test.k8c.io",
		Organization: "k8c",
		AltNames:     []string{"localhost"},
		ValidForDays: 365,
	})

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = certsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	mockCtrl = gomock.NewController(GinkgoT())
	ca = mocks.NewMockCertAuthority(mockCtrl)

	Expect((&controller.CertificateReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		CA:     ca,
	}).SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).NotTo(HaveOccurred(), "failed to start manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
