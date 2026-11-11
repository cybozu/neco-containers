# ct-monitor

1. Check [releases](https://github.com/Hsn723/ct-monitor/releases) for changes.
2. Update `ARG VERSION` in `Dockerfile`.
3. Update `E2ETEST_K8S_VERSION`, `E2ETEST_K8S_IMAGE`, `CERT_MANAGER_VERSION` in `e2e/Makefile` and kind/kubectl versions in `e2e/aqua.yaml` if necessary.
4. Run `aqua update-checksum -prune` in `e2e` if `e2e/aqua.yaml` is updated.
5. Upgrade direct dependencies in `go.mod` (for e2e test) and run `go mod tidy`.
6. Upgrade direct dependencies in `plugin/incluster-filter/go.mod` and run `go mod tidy` in `plugin/incluster-filter/`.
7. Update `TAG` file.
