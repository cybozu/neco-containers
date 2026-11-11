# ceph-csi-operator

![CSA Update](../csa_update.svg)

1. See [Rook's Chart.yaml file](https://github.com/rook/rook/blob/master/deploy/charts/rook-ceph/Chart.yaml) and [ceph-csi-operator's values.yaml file](https://github.com/ceph/ceph-csi-operator/blob/main/deploy/charts/ceph-csi-operator/values.yaml) of the appropriate tags to check the version of ceph-csi-operator.
2. Update `VERSION` in Dockerfile with the value which you checked in the previous step.
3. See the upstream Dockerfile of the appropriate tag, and update our Dockerfile if necessary. The upstream Dockerfile is [here](https://github.com/ceph/ceph-csi-operator/blob/main/Dockerfile).
4. Update `BRANCH` and `TAG` files.
