# bmc-reverse-proxy

![Kubernetes Update](../kubernetes_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Confirm test are green.

   ```bash
   make test
   ```

3. Update image tag in `bmc-reverse-proxy.yaml`.
4. Update `TAG` file.
