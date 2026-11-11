# cep-checker

![Regular Update](../regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update versions in `e2e/Makefile`, `e2e/aqua.yaml`, and `go.mod` to the versions used by neco.
   - Cilium: `CILIUM_VERSION` in `e2e/Makefile`
   - cilium-cli: version in `e2e/aqua.yaml`
   - Kubernetes (if needed): `E2ETEST_K8S_VERSION`, `E2ETEST_K8S_IMAGE` (image digest) in `e2e/Makefile`, kind and kubectl versions in `e2e/aqua.yaml`
3. Run `aqua update-checksum -prune` in `e2e` if `e2e/aqua.yaml` is updated.
4. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...
