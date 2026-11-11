# unbound_exporter

1. Check the [release page](https://github.com/letsencrypt/unbound_exporter/releases)
2. Update our forked repository. We apply a patch to the original unbound\_exporter to support `SO_REUSEPORT`. This patch helps the restart of unbound\_exporter in a Pod where `hostNetwork` is set to `true`.
    1. Set up a local copy of the unbound\_exporter repository if you have not yet.

        ```bash
        $ cd github.com/letsencrypt
        $ git clone https://github.com/letsencrypt/unbound_exporter.git
        $ cd unbound_exporter
        $ git remote add fork ssh://gh/cybozu-go/unbound_exporter
        ```

    2. Update our forked repository.

        ```bash
        $ cd github.com/letsencrypt/unbound_exporter
        $ git checkout main
        $ git pull
        $ git fetch fork
        $ git push fork main
        ```

    3. Find a commit of our patch in the old tag.

        ```bash
        $ git log vx.y.z-neco.w
        ```

    4. Apply our patch to the copy of the latest tag.

        ```bash
        $ git checkout -b neco-vX.Y.Z-patch vX.Y.Z
        $ git cherry-pick <commit>

        If it causes conflicts:
        $ (edit to resolve conflicts)
        $ git add foobar
        $ git cherry-pick --continue
        ```

    5. Test the patched branch.

        ```bash
        $ staticcheck ./...
        # staticcheck will fail due to ST1005, which points out that error strings should not be capitalized.
        # $ staticcheck --checks inherit,-ST1005 ./...
        $ go vet ./...
        $ go test ./...
        ```

    6. Push the patched branch.

        ```bash
        $ git push -u fork neco-vX.Y.Z-patch
        ```

    7. Create and push a tag.

        ```bash
        $ git tag -a -m "reuse-port" vX.Y.Z-neco.1
        $ git push fork tag vX.Y.Z-neco.1
        ```

3. Update `UNBOUND_EXPORTER_COMMIT` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
