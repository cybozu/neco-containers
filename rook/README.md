Rook container
==============

This container uses a cybozu's own Rook [cybozu-go/rook][]:neco-release. It's because the following feature and fixes are necessary for Neco and are not in the upstream's newest stable version (v1.3.0):

* Some trivial bugs about dm devices  (see the Path at [cybozu-go/rook][]:use-dmcrypt-dev)
* Update helm chart manifests (see [this][ceph: update helm version] commit)

Our custom Rook is created as follows to support the above-mentioned feature and fixes.

Until a release from Rook supports the above feature, the following update procedure is needed:

```
# Fetch updates from origin and fork
cd $GOPATH/src/github.com/rook/rook
git fetch origin
git remote add fork git@github.com:cybozu-go/rook.git
git fetch fork
git checkout neco-release

# Rebase and push to the forked repository
# Please resolve conflict & dependencies carefully
git rebase origin/master
git push -f fork neco-release

# Set $ROOK_VERSION & $MASTER_COMMIT, e.g. ROOK_VERSION="1.3.0"; MASTER_COMMIT="7701c0"
ROOK_VERSION=<LATEST STABLE VERSION>
MASTER_COMMIT=$(git rev-parse master | cut -c 1-7)
git tag "v$ROOK_VERSION-master-$MASTER_COMMIT"
git push fork "v$ROOK_VERSION-master-$MASTER_COMMIT"
```

Note that when a stable version of Rook starts to support the above-mentioned feature and fixes, please update this procedure.

[rook]: https://github.com/rook/rook
[cybozu-go/rook]: https://github.com/cybozu-go/rook
[ceph: update helm version]: https://github.com/rook/rook/commit/a86b06084988d155450557679602a5422b6e6b2c

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/rook)
