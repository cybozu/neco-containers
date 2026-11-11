# hubble

See also [the common instructions for Cilium images](../MAINTENANCE.md).

> [!Note]
> Keep the Hubble CLI on the same minor version as Cilium in each `cilium-X.Y/hubble`.

1. Check the [releases](https://github.com/cilium/hubble/releases) page for changes.
2. Update `HUBBLE_SHA` in `Dockerfile` with the commit SHA of the new release tag.
3. Update the `BRANCH` and `TAG` files accordingly.

Hubble image is no longer built by the upstream. If failing to build the image, check upstream changes.
