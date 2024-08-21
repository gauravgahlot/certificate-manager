//go:build e2e

package controller_test

import (
	"time"

	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	certsv1 "certificate-manager/api/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	secretName           = "test-secret"
	certificateName      = "test-certificate"
	certificateNamespace = "test-namespace-"

	timeout  = 10 * time.Second
	interval = time.Second
)

var _ = Describe("Certificate Controller", func() {
	var ns corev1.Namespace
	BeforeEach(func() {
		ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: certificateNamespace},
		}

		err := k8sClient.Create(ctx, &ns)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		err := k8sClient.Delete(ctx, &ns)
		Expect(err).NotTo(HaveOccurred(), "failed to delete test namespace")
	})

	Context("When creating a certificate", func() {
		It("Should be able to utilise CertAuthority", func() {
			ca.EXPECT().IssueCert(gomock.Any()).AnyTimes().Return(tlsKey, tlsCrt, nil)
			ca.EXPECT().HasCertificateExpired(gomock.Any()).AnyTimes().Return(false, nil)

			cert := &certsv1.Certificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      certificateName,
					Namespace: ns.Name,
				},
				Spec: certsv1.CertificateSpec{
					Organization: "k8c",
					DNSName:      "test.k8c.io",
					AltNames:     []string{"localhost"},
					SecretRef: certsv1.SecretRef{
						Name: secretName,
					},
				},
			}

			Expect(k8sClient.Create(ctx, cert)).Should(Succeed())
			key := types.NamespacedName{Name: certificateName, Namespace: ns.Name}
			createdCrt := &certsv1.Certificate{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, createdCrt)
				return err == nil && createdCrt.Status.State != ""
			}, timeout, interval).Should(BeTrue())

			Expect(createdCrt.Name).Should(Equal(certificateName))
			Expect(createdCrt.Spec.DNSName).Should(Equal("test.k8c.io"))
			Expect(createdCrt.Spec.ValidForDays).Should(Equal(365))
			Expect(createdCrt.Status.State).Should(Equal(certsv1.StateValid))

			Expect(k8sClient.Delete(ctx, createdCrt)).Should(Succeed())
		})
	})

	Context("When updating a certificate", func() {
		It("Should be able to utilise CertAuthority", func() {
			ca.EXPECT().IssueCert(gomock.Any()).AnyTimes().Return(tlsKey, tlsCrt, nil)
			ca.EXPECT().HasCertificateExpired(gomock.Any()).AnyTimes().Return(false, nil)

			cert := &certsv1.Certificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      certificateName,
					Namespace: ns.Name,
				},
				Spec: certsv1.CertificateSpec{
					Organization: "k8c",
					DNSName:      "test.k8c.io",
					SecretRef: certsv1.SecretRef{
						Name: secretName,
					},
				},
			}

			Expect(k8sClient.Create(ctx, cert)).Should(Succeed())

			key := types.NamespacedName{Name: certificateName, Namespace: ns.Name}
			createdCrt := &certsv1.Certificate{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, createdCrt)
				return err == nil && createdCrt.Status.State != ""
			}, timeout, interval).Should(BeTrue())

			cert.Spec.AltNames = []string{"localhost"}
			createdCrt.Spec.AltNames = []string{"localhost"}
			Expect(k8sClient.Update(ctx, createdCrt)).Should((Succeed()))

			updatedCrt := &certsv1.Certificate{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, updatedCrt)
				return err == nil && updatedCrt.Status.State != ""
			}, timeout, interval).Should(BeTrue())

			Expect(updatedCrt.Name).Should(Equal(certificateName))
			Expect(updatedCrt.Status.State).Should(Equal(certsv1.StateValid))
			Expect(updatedCrt.Name).Should(Equal(certificateName))
			Expect(len(updatedCrt.Spec.AltNames)).Should(Equal(1))
			Expect(updatedCrt.Spec.AltNames).Should(Equal([]string{"localhost"}))

			Expect(k8sClient.Delete(ctx, createdCrt)).Should(Succeed())
		})
	})
})
