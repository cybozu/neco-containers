# kube-storage-version-migrator

1. Check the [release page](https://github.com/kubernetes-sigs/kube-storage-version-migrator/releases).
2. Check the upstream build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/Makefile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/initializer/Dockerfile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/migrator/Dockerfile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/trigger/Dockerfile`
3. Update `MIGRATOR_VERSION` and `MIGRATOR_COMMIT` in `Dockerfile`
4. Update `TAG` file.
