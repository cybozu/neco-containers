# coredns

1. Check the [release page](https://github.com/coredns/coredns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/coredns/coredns/blob/vX.Y.Z/Dockerfile`
3. Update `COREDNS_COMMIT` and the comment with the link to the release page (`https://github.com/coredns/coredns/releases/tag/vX.Y.Z`) in `Dockerfile`.
   - Use the commit SHA that the tag points to, not the SHA of the annotated tag object (`git ls-remote --tags https://github.com/coredns/coredns 'refs/tags/vX.Y.Z*'`).
4. Update the base images in `Dockerfile`.
   - Build stage: the latest `ghcr.io/cybozu/golang` image that satisfies the Go version required by coredns (`go` and `toolchain` in `go.mod`).
   - Runtime stage: the latest `ghcr.io/cybozu/ubuntu` image.
   - Pin them by the full tag and digest (e.g. `ghcr.io/cybozu/golang:X.Y.Z.N_resolute@sha256:...`).
5. Update `BRANCH` and `TAG` files.
