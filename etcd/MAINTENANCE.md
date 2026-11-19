# etcd

1. Check the [release page](https://github.com/etcd-io/etcd/releases).
2. Check the upstream Dockerfile `https://github.com/etcd-io/etcd/blob/vX.Y.Z/Dockerfile`.
   - If there are any updates, update our `Dockerfile`.
3. Update `ETCD_COMMIT` and the release page link in `Dockerfile`.
   - Click the commit link next to the tag on the release page to get the full commit SHA.
   - You can also get the commit SHA with `git ls-remote https://github.com/etcd-io/etcd 'refs/tags/vX.Y.Z^{}'`.
4. Update the base images in `Dockerfile` to the latest ones based on the latest Ubuntu version.
   - Build stage: Update to the latest [`ghcr.io/cybozu/golang`](https://github.com/cybozu/neco-containers/pkgs/container/golang) image that satisfies the Go version in `toolchain` of `go.mod` in the etcd repository.
   - Runtime stage: Update to the latest [`ghcr.io/cybozu/ubuntu`](https://github.com/cybozu/ubuntu-base/pkgs/container/ubuntu) image.
5. Pin the base images by tag and digest.
   - You can find the digest of each tag on the image pages linked above.
   - You can also get the digest with `docker buildx imagetools inspect <image>:<tag>`.
6. Update `TAG`.
7. Update `BRANCH`.
