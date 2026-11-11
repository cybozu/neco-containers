# cilium-operator-generic

See also [the common instructions for Cilium images](../MAINTENANCE.md).

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/operator/Dockerfile`

> [!Note]
> `ENV CILIUM_SHA` in `neco-containers/cilium-X.Y/cilium-operator-generic/Dockerfile` pins a specific commit SHA of the `cybozu-go/cilium` fork branch (e.g. `vX.Y.Z-cybozu`), not the branch tip.
> Changes to that branch must go through a pull request. After merging one (e.g. additional cherry-picked patches), you must update `CILIUM_SHA` to the new commit SHA; it will not be picked up automatically.
> `make check-cilium-target` (run by CI before the image build) fails if `CILIUM_SHA` is not the current tip of that branch, to catch this case.

3. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The cilium-operator-generic image should be updated at the same time as the [cilium](../cilium/MAINTENANCE.md) image in the same `cilium-X.Y/` directory for consistency.
