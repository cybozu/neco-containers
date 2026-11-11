# neco-exporter

1. Update the software versions in the following files to the ones used by neco.
   - `Dockerfile`
   - `Makefile`
   - `e2e/Makefile`
   - `go.mod`
   - `e2e/aqua.yaml`
2. Run `aqua update-checksum -prune` in `e2e` if `e2e/aqua.yaml` is updated.
3. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...
