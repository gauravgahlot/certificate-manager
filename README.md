# Certificate Manager

Certificate Manager is a Kubernetes controller which enables developers to
request TLS certificates that they can incorporate into their application
deployments.

Developers, while deploying their applications, are required to include a custom
`Certificate` resource with their application manifest.

The certificate-manager watches the `Certificate` custom resource, processes it
and creates matching TLS certificate secrets. The certificates issued by the
certificate-manager are **self-signed**.

## Table of Contents

- [Architecture](./docs/architectue.md)
- [Canonical Defintion of a Certificate](./docs/architectue.md/#canonical-definition-of-a-certificate)
- [How to use it?](./docs/how-to-use.md)
- [Demo](#demo)
- [References](#references)

## Demo

[![asciicast](https://asciinema.org/a/RtXnx6uKhp2cfSDVZKHLCollS.svg)](https://asciinema.org/a/RtXnx6uKhp2cfSDVZKHLCollS)

You can quickly deploy and test the certificate manager using the followig steps.
However, please note that the setup assumes you are using a
[Kind](https://kind.sigs.k8s.io/) cluster.

```sh
# connect to an existing Kind cluster
export KUBECONFIG=~/.kube/kind.yaml

# (optional) create a new kind cluster
kind create cluster

# build and push Docker image for 'certificate-manager' and 'todo-app'
make docker-build docker-push

# install the manifests to the cluster, and start the controller manager
make install

# deploy the `todo-app` to the cluster and run the test script
make test-app

# (optional) run e2e tests
make e2e
```

## References

- https://pkg.go.dev/crypto/x509
- https://book.kubebuilder.io/
- https://github.com/kubernetes-sigs/controller-runtime/blob/main/examples/crd/
