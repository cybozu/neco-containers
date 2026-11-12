# hubble-relay

See also [the common instructions for Cilium images](../MAINTENANCE.md).

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/hubble-relay/Dockerfile`
3. Update `CILIUM_SHA` in `Dockerfile`.
4. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The hubble-relay image should be updated at the same time as the [cilium](../cilium/MAINTENANCE.md) image in the same `cilium-X.Y/` directory for consistency.
