# squid-exporter

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update squid version in `Makefile` and `e2e/pod.yaml` if there are any updates.
3. Update `E2ETEST_K8S_VERSION`, `E2ETEST_K8S_IMAGE` (image digest) in `e2e/Makefile` and kind/kubectl versions in `e2e/aqua.yaml` if necessary.
4. Run `aqua update-checksum -prune` in `e2e` if `e2e/aqua.yaml` is updated.
5. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

> [!Note]
> The squid images should be updated at the same time as the squid-exporter image for consistency.
