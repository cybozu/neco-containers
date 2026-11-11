# local-pv-provisioner

![CSA Update](../csa_update.svg)

1. Update version variables in `Makefile` and `e2etest/Makefile`.
   - `Makefile`: `CONTROLLER_TOOLS_VERSION` and `ENVTEST_K8S_VERSION`
   - `e2etest/Makefile`: `KUBERNETES_VERSION` and `MINIKUBE_VERSION`
2. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
   1. `go get -u`
   1. `VERSION=0.x.y; go get k8s.io/api@v${VERSION} k8s.io/apimachinery@v${VERSION} k8s.io/client-go@v${VERSION}`
   1. `VERSION=1.x.y; go get github.com/prometheus/client_golang@v${VERSION}`
      - Use the same version as Kubernetes's go.mod (`https://github.com/kubernetes/kubernetes/blob/<upgrading Kubernetes release version>/go.mod`).
   1. `VERSION=1.x.y; go mod tidy -go=${VERSION}`
      - Use the same version as Kubernetes's go.mod (`https://github.com/kubernetes/kubernetes/blob/<upgrading Kubernetes release version>/go.mod`).
3. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/local-pv-provisioner
   make generate manifests
   # Commit, if there are any updated files.
   ```

4. Confirm build and test are green.

   ```bash
   make build test
   ```

5. Update `TAG` file.
