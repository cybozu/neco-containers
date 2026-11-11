# spire-server

1. Check the [release page](https://github.com/spiffe/spire/releases).
2. Check the upstream Dockerfile and Makefile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/spiffe/spire/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/spiffe/spire/blob/vX.Y.Z/Makefile`
3. Update the base image in `Dockerfile`. Pin it by the full tag and digest (e.g. `ghcr.io/cybozu/golang:X.Y.Z.N_noble@sha256:...`).
4. Update `TAG` file.

> [!Note]
> The spire-server, spire-agent, and spiffe-csi-driver images should be updated at the same time for consistency.
