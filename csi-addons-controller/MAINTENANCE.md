# csi sidecars/csi-addons-controller

This section applies to the following containers. These containers are maintained similarly.

- csi-addons-controller
- csi-addons-sidecar
- csi-attacher
- csi-node-driver-registrar
- csi-provisioner
- csi-resizer
- csi-snapshotter

1. See [Rook's values.yaml file](https://github.com/rook/rook/blob/master/deploy/charts/rook-ceph/values.yaml) of the appropriate tag and check the version of csi sidecars. The csi-addons-controller version should be the same as the csi-addons-sidecar version.
2. Update `VERSION` in Dockerfile with the value which you checked in the previous step.
3. See the upstream Dockerfile of the appropriate tag, and update our Dockerfile if necessary. The upstream Dockerfile is listed below.
   - [csi-addons-controller](https://github.com/csi-addons/kubernetes-csi-addons/blob/main/Dockerfile)
   - [csi-addons-sidecar](https://github.com/csi-addons/kubernetes-csi-addons/blob/main/build/Containerfile.sidecar)
   - [csi-attacher](https://github.com/kubernetes-csi/external-attacher/blob/master/Dockerfile)
   - [csi-node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar/blob/master/Dockerfile)
   - [csi-provisioner](https://github.com/kubernetes-csi/external-provisioner/blob/master/Dockerfile)
   - [csi-resizer](https://github.com/kubernetes-csi/external-resizer/blob/master/Dockerfile)
   - [csi-snapshotter](https://github.com/kubernetes-csi/external-snapshotter/blob/master/cmd/csi-snapshotter/Dockerfile)
4. update image tag in `Dockerfile` if necessary.
5. Update `BRANCH` and `TAG` files.

> [!Note]
> You can choose the latest stable Ubuntu image for runtime. The upstream uses distroless as the base image for runtime, while Neco uses Ubuntu for easier debugging.
> We have not decided to use Ubuntu for csi-addons-controller and csi-addons-sidecar yet, therefore, they use distroless as the base image.

<br>

> [!Note]
> You may choose the latest docker image for the build, regardless of the upstream go version. The current go compiler builds with the language version and toolchain version based on the go version specified in the go.mod file. There is no need to use an older version of the image to match go.mod. As a known issue, the upstream build script warns that `test-gofmt and test-vendor are known to be sensitive to the version of Go.`. However, we use the latest docker image unless the test fails.
