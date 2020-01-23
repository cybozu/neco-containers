[![Docker Repository on Quay](https://quay.io/repository/cybozu/rook/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/rook)

Rook container
==============

This container uses a cybozu's own Rook [cybozu-go/rook:release][]. It's because the following features are necessary for Neco and are not in the upstream's newest stable version (v1.2.2):

* Supporting partition device type (`blkid`'s "part" type), which is only enabled in `rook/rook:master`.
* Supporting `topologySpreadConstraints`. We created a branch to support it. However, merging it to upstream Rook would take time (see [Rook's issue](https://github.com/rook/rook/issues/4387)).

Our custom Rook is created as follows to support the above-mentioned features.

Until a release from Rook supports the above two features, the following update procedure is needed:

```
# Please set $ROOK_VERSION & $MASTER_COMMIT, e.g. ROOK_VERSION="1.2.2"; MASTER_COMMIT="a42d822"
cd go/src/github.com/cybozu-go/rook
git remote add upstream git@github.com:rook/rook.git
git checkout neco-release && git pull
git fetch upstream
git rebase upstream/master
# Please resolve conflict & dependencies carefully

git push -f origin neco-release
git tag "v$ROOK_VERSION-master-$MASTER_COMMIT"
git push origin "v$ROOK_VERSION-master-$MASTER_COMMIT"
```

After that, please set `TAG` as `$ROOK_VERSION-master` and increment `BRANCH` (i.e. the image tag will be labeled as `v$ROOK_VERSION-master.$BRANCH`). And please update Dockerfile to use the tag `v$ROOK_VERSION-master-$MASTER_COMMIT`.

Note that when a stable version of Rook starts to support the above-mentioned features, please update this procedure.

[rook]: https://github.com/rook/rook
