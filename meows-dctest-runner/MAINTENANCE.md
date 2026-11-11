# meows-dctest-runner

The directory is split into per-OS subdirectories (`meows-dctest-runner/ubuntu22.04/`, `meows-dctest-runner/ubuntu24.04/`). Repeat the following steps for each subdirectory.

1. Check the [release page](https://github.com/cybozu-go/meows/releases) and [runner image release](https://github.com/cybozu-go/meows/pkgs/container/meows-runner).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/runner-images/ubuntu22.04/Dockerfile` (for `ubuntu22.04/`)
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/runner-images/ubuntu24.04/Dockerfile` (for `ubuntu24.04/`)
3. Update `MEOWS_RUNNER_TAG` in `Dockerfile`.
4. Update `GO_VERSION` and `PLACEMAT_VERSION` in `Dockerfile`, if there are any updates.
   - `GO_VERSION`: <https://github.com/cybozu/neco-containers/blob/main/golang-all>
   - `PLACEMAT_VERSION`: <https://github.com/cybozu-go/placemat/releases/latest>
5. Update `BRANCH` and `TAG` files.
