# cephcsi

1. See [Rook's values.yaml file](https://github.com/rook/rook/blob/master/deploy/charts/rook-ceph/values.yaml) of the appropriate tag and check the version of cephcsi.
2. Update `CSI_IMAGE_VERSION` in Dockerfile with the value which you checked in the previous step.
3. Update `BASE_IMAGE` in Dockerfile if necessary.
   - If `BASE_IMAGE` is too old, the build may fail.
   - You should also check `BASE_IMAGE` in [the upstream build.env](https://github.com/ceph/ceph-csi/blob/devel/build.env) file of the appropriate tag.
   - If you are going to update Rook later, you should also check `CEPH_VERSION` in [the Rook's upstream Makefile](https://github.com/rook/rook/blob/master/images/ceph/Makefile) file of the appropriate tag.
   - If the desired version of ceph image does not exist in our container repository, build the ceph image first.
4. See [the upstream Dockerfile](https://github.com/ceph/ceph-csi/blob/devel/deploy/cephcsi/image/Dockerfile) of the appropriate tag, and update our Dockerfile if necessary.
5. Update `BRANCH` and `TAG` files.

> [!Note]
> Because cephcsi container is build based on the ceph container, build the ceph container first if necessary.
