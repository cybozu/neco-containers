# pause

1. Check the changelog `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/CHANGELOG.md`.
2. Check the upstream Dockerfile `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/Dockerfile`.
   - If there are any updates, update our `Dockerfile`.
3. Update `K8S_COMMIT` and the release page link in `Dockerfile`.
   - Click the commit link next to the tag on the kubernetes release page to get the full commit SHA.
   - You can also get the commit SHA with `git ls-remote https://github.com/kubernetes/kubernetes 'refs/tags/vX.Y.Z^{}'`.
4. Update `PAUSE_VERSION` in `Dockerfile` to the latest version in the changelog.
5. Update the build stage image to the latest [`ghcr.io/cybozu/ubuntu-dev`](https://github.com/cybozu/ubuntu-base/pkgs/container/ubuntu-dev) image based on the latest Ubuntu version.
6. Pin the build stage image by tag and digest.
   - You can find the digest of each tag on the image page linked above.
   - You can also get the digest with `docker buildx imagetools inspect <image>:<tag>`.
7. Update `TAG`.
8. Update `BRANCH`.
