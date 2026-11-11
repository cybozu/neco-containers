# ceph-extra-exporter

![CSA Update](../csa_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Upgrade base images in `Dockerfile`.
3. Update the tool versions in `aqua.yaml` if necessary.
4. Run `aqua update-checksum -prune` if `aqua.yaml` is updated.
5. Update the `TAG` files accordingly.
