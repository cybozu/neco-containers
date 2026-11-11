# cert-manager

1. Check [releases](https://github.com/jetstack/cert-manager/releases) for changes.
2. Check whether manually applied patches have been included in the new release and remove them accordingly.
   1. If patches are still needed, synchronize the forked repository (<https://github.com/cybozu-go/cert-manager>).
   2. Create and checkout a new branch named `vX.Y.Z-neco` from the tag named `vX.Y.Z`.
   3. Cherry-pick the commit included patches and create a new tag named `vX.Y.Z-neco.1`.
   4. Push it.
3. Update base images in `Dockerfile`.
4. Update `CERT_MANAGER_SHA` in `Dockerfile`.
5. Update image tag in `README.md`.
6. Update `BRANCH` and `TAG` files.
