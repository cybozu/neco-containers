# cilium-1.18

Cilium images are split into per-minor-version directories (e.g. `cilium-1.18/`).
Each `cilium-X.Y/` directory contains the images that must match the Cilium minor version:

- [`cilium-X.Y/cilium`](./cilium/MAINTENANCE.md)
- [`cilium-X.Y/cilium-operator-generic`](./cilium-operator-generic/MAINTENANCE.md)
- [`cilium-X.Y/hubble-relay`](./hubble-relay/MAINTENANCE.md)
- [`cilium-X.Y/hubble`](./hubble/MAINTENANCE.md)

[cilium-certgen](../cilium-certgen/MAINTENANCE.md) and [hubble-ui](../hubble-ui/MAINTENANCE.md) are shared by all Cilium versions and live in top-level directories.

To update an existing minor version, follow the `MAINTENANCE.md` of each image inside `cilium-X.Y/`.

To add a new minor version:

1. Copy the latest `cilium-X.Y/` directory to `cilium-X.Y'/`.
2. Update the four images following each `MAINTENANCE.md`. Update the `BRANCH` (`X.Y'`) and `TAG` files of all of them.
3. Update `scan.identifier` in `cilium/build-targets.yaml` and `hubble-relay/build-targets.yaml` to `ghcr.io/cybozu/<image>:X.Y'`.
4. Add `/cilium-X.Y'/` to `.github/CODEOWNERS`.

To remove an old minor version, check that no cluster uses it, then delete `cilium-X.Y/` and its CODEOWNERS entry.

> [!Note]
> Changes that are not specific to a Cilium version (e.g. pinning base images, CI-related settings) should be applied to all `cilium-X.Y/` directories in the same PR.

> [!Note]
> The cilium-operator-generic and hubble-relay images should be updated at the same time as the cilium image in the same `cilium-X.Y/` directory for consistency.
