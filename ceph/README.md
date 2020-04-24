[![Docker Repository on Quay](https://quay.io/repository/cybozu/ceph/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/ceph)

Ceph container
==============

Build Docker container image for [Ceph][], a distributed object, block, and file storage platform.

Usage
-----

This container image assumes to be used by Rook.
To use in Rook, you need to write a manifest of the custom resource CephCluster with this image.

Reference for updating image
----------------------------

This container uses a cybozu's patched Ceph [cybozu/ceph][]:neco-release. It's because using dmcrypt devices that is already encrypted by users is not supported in the upstream's newest stable version (v15.2.1), but the function is  necessary for Neco.

Our custom Ceph is created as follows to support the above-mentioned feature.

Until a release from Ceph supports the above feature, the following update procedure is needed:

```
# Please set $CEPH_VERSION, e.g. CEPH_VERSION="14.2.8.4"
cd go/src/github.com/ceph/ceph
git checkout master && git pull
git remote add fork git@github.com:cybozu/ceph.git
git fetch fork && git checkout neco-release
git rebase master
# Please resolve conflict & dependencies carefully

git push -f fork neco-release
git tag -a "v$CEPH_VERSION"
git push fork "v$CEPH_VERSION"
```

Note that when a stable version of Ceph starts to support the above-mentioned feature and fixes, please update this procedure.

[Ceph]: https://github.com/ceph/ceph
[cybozu/ceph]: https://github.com/cybozu/ceph
