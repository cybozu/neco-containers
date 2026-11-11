# rook

![CSA Update](../csa_update.svg)

> [!Note]
> A specific version of rook depends on specific versions of csi sidecar containers listed below. Update these containers at the same time.
> 
> - cephcsi
> - ceph-csi-operator
> - csi-attacher
> - csi-node-driver-registrar
> - csi-provisioner
> - csi-resizer
> - csi-snapshotter
> - csi-addons-controller
> - csi-addons-sidecar

1. Check the [release page](https://github.com/rook/rook/releases).
2. Check the `CEPH_VERSION` of the upstream Makefile. If the version is changed, update the base image version in our `Dockerfile`.
   - `https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Makefile`
   - If the desired version of ceph image does not exist in our container repository, build the ceph image first.
3. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Dockerfile`
4. Update build image tag in Dockerfile if necessary.
5. Update `ROOK_VERSION` in `Dockerfile`.
6. Update `BRANCH` and `TAG` files.

> [!Note]
> You may choose the latest docker image for the build, regardless of the upstream go version.
> The current go compiler builds with the language version and toolchain version based on the go version specified in the go.mod file.
> There is no need to use an older version of the image to match go.mod.
