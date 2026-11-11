# coredns

![Kubernetes Update](../kubernetes_update.svg)

1. Check the [release page](https://github.com/coredns/coredns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/coredns/coredns/blob/vX.Y.Z/Dockerfile`
3. Update `COREDNS_COMMIT` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.
