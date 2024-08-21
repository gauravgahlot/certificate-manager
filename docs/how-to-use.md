## How to use it?

Suppose you have an application such as [todo-app](../todo-app/). It's a web
server that serves a single endpoint `/todo`. The server expects a TLS key and
certificate to be present, so that it can securely serve the requests.

However, you do not want to manage the TLS certificate by yourself. And that's
where Certificate Manager comes in.

All you have to do is create a `Certificate` request by submitting a manifest
similar to the one below:

```yaml
apiVersion: certs.k8c.io/v1
kind: Certificate
metadata:
  name: todo-app
  namespace: todo
spec:
  dnsName: todo-app.todo.svc.cluster.local
  organization: k8c
  validForDays: 90
  altNames:
    - localhost
    - todo-app
  secretRef:
    name: todo-app
```

You can read more about the canonical definition of a `Certificate` in
[this document](./architectue.md/#canonical-definition-of-a-certificate).

When the request is submitted to Kubernetes API server, the certificate manager
will generate a self-signed TLS certificate using the provided details. The
certificate is then stored in a Kubernetes `Secret` (type `TLS`) named as `todo-app`.

As a next step, you need to use the `Secret`, `todo-app` in this case, in your
application deployment as shown below:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: todo-app
  namespace: todo
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: todo-app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: todo-app
    spec:
      containers:
        - name: todo-app
          image: todo-app:v0.1.0
          ports:
            - containerPort: 443
          volumeMounts:
            - name: certs
              mountPath: "/tmp/certs"
              readOnly: true
      volumes:
        - name: certs
          secret:
            secretName: todo-app
```

The above spec will mount the `todo-app` Secret as a volume in the application Pod
at `/tmp/certs`.

As a final step, we need to read the TLS key and certifcate to secure our server:

```go
// certificate-manager/todo-app/main.go

const port = ":443"
var (
	certFile = filepath.Join("tmp", "certs", "tls.crt")
	keyFile  = filepath.Join("tmp", "certs", "tls.key")
)

func main() {
	// register the handler for /todo path
	http.HandleFunc("/todo", handleTodo)

	slog.Info("starting HTTP server", "port", port)
	err := http.ListenAndServeTLS(port, certFile, keyFile, nil)
	if err != nil {
		slog.Error("server failure", "error", err)
	}
}
```

The complete code can be found at [todo-app/main.go](../todo-app/main.go).

In order to test if everything is working as expected, follow and execute
the steps below from the root of this repository:

```sh
# connect to an existing Kind cluster
export KUBECONFIG=~/.kube/kind.yaml

# (optional) create a new kind cluster
kind create cluster

# build and push Docker image for 'certificate-manager' and 'todo-app'
make docker-build docker-push

# install the manifests to the cluster, and start the controller manager
make install

# deploy the todo-app to the cluster
kubectl apply -f todo-app/deploy.yaml
```

Create a port-forward for the todo-app service:

```sh
kubectl port-forward -n todo svc/todo-app 8443:443 &
[1] 65387
Forwarding from 127.0.0.1:8443 -> 443
Forwarding from [::1]:8443 -> 443
```

Let's first try an _insecure_ connection to the server using `curl` with `-k` option:

```sh
curl -sk https://localhost:8443/todo

[{"dueDate":"2024-08-28T06:31:38.031982924Z","id":"774ab9a4-5b72-4bee-bbc2-05980f0283a1","title":"write a todo-app"},{"dueDate":"2024-08-29T06:31:38.032002758Z","id":"85edeafb-00d7-49fe-92d6-58c075762688","title":"define K8s manifests"},{"dueDate":"2024-08-30T06:31:38.032004466Z","id":"3406005f-7e4e-4847-99aa-95930c48694b","title":"use certificates"}]
```

Great, the sever is responding.

As a final test, we want to connect with the server over a _secure_ connection.

```sh
curl  https://localhost:8443/todo

Handling connection for 8443
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.
E0821 12:05:23.688226   65387 portforward.go:394] error copying from local connection to remote stream: read tcp6 [::1]:8443->[::1]:64944: read: connection reset by peer
```

_What's wrong?_

As the error message suggests, `curl` is unable to verify the legitimacy of the server
and therefore, could not establish a secure connection to it.

Let's fix it by obtaining the server certificate, issued by the certificate manager, and
stored in the `todo-app` secret:

```sh
kubectl get secret -n todo todo-app -o jsonpath='{.data.tls\.crt}' | base64 -d > tls.crt
```

Now, rerun the `curl` command with the `--cacert` option:

```sh
curl --cacert tls.crt https://localhost:8443/todo

[{"dueDate":"2024-08-28T06:39:34.585780298Z","id":"fd8133f1-ab1d-4e3f-8e92-e5f74793e7ea","title":"write a todo-app"},{"dueDate":"2024-08-29T06:39:34.585792923Z","id":"24f68065-9195-45df-85cd-38e8fc810bd5","title":"define K8s manifests"},{"dueDate":"2024-08-30T06:39:34.585795006Z","id":"ae7aec7c-c9b9-4633-b169-6ef96378e7df","title":"use certificates"}]
```

Congratulations!! You have successfully setup a secure connection with the server.
