[![Docker Repository on Quay](https://quay.io/repository/cybozu/rook/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/rook)

Rook container
==============

This directory provides a Dockerfile to build a rook container that runs [rook][].

Because of the following reasons, the Dockerfile uses [cybozu-go/rook:release][].
1. The current version of Rook (v1.2.2) does not support partition (`part`) device type, though `master` branch already supported it. So we need to use `master` branch source code.
2. The current version of Rook (v1.2.2) does not support `topologySpreadConstraints` feature. So we need to patch the original code of Rook.

Until a release from Rook supports the above two features, the following update procedure is needed:

```
# Please set $ROOK_VERSION & $MASTER_COMMIT, e.g. ROOK_VERSION="1.1.2"; MASTER_COMMIT="a42d822"
cd go/src/github.com/cybozu-go/rook
git remote add upstream git@github.com:rook/rook.git
git checkout neco-release && git pull
git rebase upstream/master
git push -f origin neco-release
git tag "v$ROOK_VERSION-master-$MASTER_COMMIT"
git push origin "v$ROOK_VERSION-master-$MASTER_COMMIT"
```

After that, please set `TAG` as `$ROOK_VERSION-master` and increment `BRANCH` (i.e. the image tag will be labeled as `v$ROOK_VERSION-master.$BRANCH`). And please update Dockerfile to use the tag `v$ROOK_VERSION-master-$MASTER_COMMIT`.

Note that when a future release from Rook supports the above two features, please modify the update procedure appropriately.

[rook]: https://github.com/rook/rook
