# kube-state-metrics

1. Check the [release page](https://github.com/kubernetes/kube-state-metrics/releases).
2. Check the upstream Dockerfile `https://github.com/kubernetes/kube-state-metrics/blob/vX.Y.Z/Dockerfile`.
   - If there are any updates, update our `Dockerfile`.
3. Update `KUBE_STATE_METRICS_COMMIT` and the release page link in `Dockerfile`.
   - Click the commit link next to the tag on the release page to get the full commit SHA.
   - You can also get the commit SHA with `git ls-remote https://github.com/kubernetes/kube-state-metrics 'refs/tags/vX.Y.Z'`.
4. Update the build stage image in `Dockerfile` to the latest one based on the latest Ubuntu version.
   - Use the latest [`ghcr.io/cybozu/golang`](https://github.com/cybozu/neco-containers/pkgs/container/golang) image that satisfies the Go version in `.go-version` of the kube-state-metrics repository.
   - The runtime stage uses `scratch`, so there is nothing to update.
5. Pin the build stage image by tag and digest.
   - You can find the digest of each tag on the image page linked above.
   - You can also get the digest with `docker buildx imagetools inspect <image>:<tag>`.
6. Update `TAG`.
