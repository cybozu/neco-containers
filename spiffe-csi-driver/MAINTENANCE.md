# spiffe-csi-driver

![Regular Update](../regular_update.svg)

1. Check the [release page](https://github.com/spiffe/spiffe-csi/releases).
2. Check the upstream Dockerfile and Makefile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/spiffe/spiffe-csi/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/spiffe/spiffe-csi/blob/vX.Y.Z/Makefile`
3. Update `TAG` file.

> [!Note]
> The spire-server, spire-agent, and spiffe-csi-driver images should be updated at the same time for consistency.
