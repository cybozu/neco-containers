[![Docker Repository on Quay](https://quay.io/repository/cybozu/rook/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/rook)

Rook container
==============

This container uses a cybozu's own Rook [cybozu-go/rook][]:neco-release. It's because the following feature and fixes are necessary for Neco and are not in the upstream's newest stable version (v1.2.6):

* Supporting `topologySpreadConstraints`. We created a branch to support it. However, merging it to upstream Rook would take time (see [Rook's issue](https://github.com/rook/rook/issues/4387)).
* Some trivial bugs and the regression about LV-backed PVC (see the Path at [cybozu-go/rook][]:use-dmcrypt-dev)

Our custom Rook is created as follows to support the above-mentioned feature and fixes.

Until a release from Rook supports the above feature, the following update procedure is needed:

```
# Please set $ROOK_VERSION & $MASTER_COMMIT, e.g. ROOK_VERSION="1.2.6"; MASTER_COMMIT="da3bf49b"
cd go/src/github.com/rook/rook
git checkout master && git pull
git remote add fork git@github.com:cybozu-go/rook.git
git fetch fork && git checkout neco-release
git rebase upstream/master
# Please resolve conflict & dependencies carefully

git push -f fork neco-release
git tag "v$ROOK_VERSION-master-$MASTER_COMMIT"
git push fork "v$ROOK_VERSION-master-$MASTER_COMMIT"
```

After that, please set `TAG` as `$ROOK_VERSION-master` and increment `BRANCH` (i.e. the image tag will be labeled as `v$ROOK_VERSION-master.$BRANCH`). And please update Dockerfile to use the tag `v$ROOK_VERSION-master-$MASTER_COMMIT`.

Note that when a stable version of Rook starts to support the above-mentioned feature and fixes, please update this procedure.

[rook]: https://github.com/rook/rook
[cybozu-go/rook]: https://github.com/cybozu-go/rook
