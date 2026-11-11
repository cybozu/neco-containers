# meows-neco-runner

![Regular Update](../regular_update.svg)

The directory is split into per-OS subdirectories (`meows-neco-runner/ubuntu22.04/`, `meows-neco-runner/ubuntu24.04/`). Repeat the following steps for each subdirectory.

1. Check the [release page](https://github.com/cybozu-go/meows/releases) and [runner image release](https://github.com/cybozu-go/meows/pkgs/container/meows-runner).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/runner-images/ubuntu22.04/Dockerfile` (for `ubuntu22.04/`)
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/runner-images/ubuntu24.04/Dockerfile` (for `ubuntu24.04/`)
3. Update the `Dockerfile`.
   - Update `GRPCURL_VERSION`, if there are any changes.
   - Update `COMMON_PACKAGES` to match the packages installed in ubuntu-debug.
     - <https://github.com/cybozu/ubuntu-base/blob/main/22.04/ubuntu-debug/Dockerfile#L6> (for `ubuntu22.04/`)
     - <https://github.com/cybozu/ubuntu-base/blob/main/24.04/ubuntu-debug/Dockerfile#L6> (for `ubuntu24.04/`)
4. Update `MEOWS_RUNNER_TAG` in `Dockerfile`.
5. Update `BRANCH` and `TAG` files.
