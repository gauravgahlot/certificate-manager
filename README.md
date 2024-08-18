# Certificate Manager

Certificate Manager is a Kubernetes controller which enables developers to
request TLS certificates that they can incorporate into their application
deployments.

Developers, while deploying their applications, are required to include a custom
`Certificate` resource with their application manifest.

The certificate-manager watches the `Certificate` custom resource, processes it
and creates matching TLS certificate secrets. The certificates issued by the
certificate-manager are **self-signed**.

## References

- https://github.com/kubernetes-sigs/controller-runtime/blob/main/examples/crd/
- https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
- https://book.kubebuilder.io/reference/markers/crd-validation
- https://pkg.go.dev/crypto/x509
