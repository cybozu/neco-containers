# kubernetes

The kubernetes directory is split into per-minor-version subdirectories named `kubernetes/k8s-X.Y/`.

Since 1.36, `BRANCH` files are not used, so the floating `X.Y` tag is not pushed.
Refer to images by the full tag in `TAG`, e.g. `ghcr.io/cybozu/kubernetes:1.36.6.1`, not `ghcr.io/cybozu/kubernetes:1.36`.

To update an existing minor version:

1. Check the [release page](https://github.com/kubernetes/kubernetes/releases).
2. Update `K8S_COMMIT` and the release page link in `Dockerfile`.
   - Click the commit link next to the tag on the release page to get the full commit SHA.
   - You can also get the commit SHA with `git ls-remote https://github.com/kubernetes/kubernetes 'refs/tags/vX.Y.Z^{}'`.
3. Update the base images in `Dockerfile` to the latest ones based on the latest Ubuntu version.
   - Build stage: Update to the latest [`ghcr.io/cybozu/golang`](https://github.com/cybozu/neco-containers/pkgs/container/golang) image that satisfies the Go version in `.go-version` of the kubernetes repository.
   - Runtime stage: Update to the latest [`ghcr.io/cybozu/ubuntu`](https://github.com/cybozu/ubuntu-base/pkgs/container/ubuntu) image.
4. Pin the base images by tag and digest.
   - You can find the digest of each tag on the image pages linked above.
   - You can also get the digest with `docker buildx imagetools inspect <image>:<tag>`.
5. If the Ubuntu version changes, make sure the packages installed in the runtime stage are still available.
6. Update `TAG`.
7. Update `BRANCH` if it exists.
8. Make sure the patch files still apply to the new version.

To add a new minor version:

1. Copy the latest minor version directory to `kubernetes/k8s-X.Y/`.
2. Update `scan.identifier` in `build-targets.yaml` to `ghcr.io/cybozu/kubernetes:X.Y`.
   - The CI job that builds the image is generated from `build-targets.yaml` automatically, so you do not need to change the GitHub Actions workflows.
3. Set `TAG` to `X.Y.Z.1`.
   - Do not create a `BRANCH` file.
4. Remove the oldest minor version directory so that only the latest 3 remain.
5. Update the rest in the same way as updating an existing minor version.
