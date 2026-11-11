# cilium

See also [the common instructions for Cilium images](../MAINTENANCE.md).

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Update `CILIUM_IMAGE_TOOLS_TARGET` in `cilium-X.Y/cilium/Makefile`. This is a commit SHA of <https://github.com/cilium/image-tools> .
   1. Run `make IMAGE_TAG=quay.io/cilium/cilium:vX.Y.Z test` to see tools versions upstream uses.
   2. See [history](https://github.com/cilium/image-tools/commits/master/) of `image-tools` to find an appropriate SHA, that provides matching tools version (especially for LLVM/Clang).
3. Checkout `cilium/cilium` and `cilium/image-tools` at the relevant SHA.
   1. Run `make checkout` and download them under `cilium-X.Y/cilium/src`.

> [!Note]
> `CILIUM_TARGET` in `cilium-X.Y/cilium/Makefile` pins a specific commit SHA of the `cybozu-go/cilium` fork branch (e.g. `vX.Y.Z-cybozu`), not the branch tip.
> Changes to that branch must go through a pull request. After merging one (e.g. additional cherry-picked patches), you must update `CILIUM_TARGET` to the new commit SHA; it will not be picked up automatically.

4. Check the upstream `Dockerfile`s to make necessary changes for `cilium-X.Y/cilium`.
   1. Run `make urls`. It displays all the URLs of the upstream `Dockerfile`s.
   2. All the build specification is written in `cilium-X.Y/cilium/Dockerfile`. Please check the header comment of the file to find the mapping of our build targets and the upstream ones.
   3. Update github.com/go-delve/delve/cmd/dlv version. Upstream may be using @latest but we must specify a version.
5. Update the cilium-cli version used by the e2e test. See cilium-cli's [README](https://github.com/cilium/cilium-cli?tab=readme-ov-file#releases) to make sure that the cilium-cli version is compatible with the cilium version.
   - Update `cilium/cilium-cli` in `aqua.yaml`, then run `aqua update-checksum -prune`.
6. Build `ghcr.io/cybozu/cilium` and see the result.
   1. Run `make build` to build.
   2. Run `make test` to make sanity check.
   3. Run `make test-e2e` to run the standard connectivity test.
      - `check-log-errors` has often been disabled to work around minor known issues.
      Check the logs for significant errors manually and consider re-enabling it.
   4. If any problem found, `dive ghcr.io/cybozu/cilium:$(cat TAG)` will help.

> [!Note]
> The cilium-operator-generic and hubble-relay images should be updated at the same time as the cilium image in the same `cilium-X.Y/` directory for consistency.
