# kubernetes

![Kubernetes Update](../kubernetes_update.svg)

The kubernetes directory is split into per-minor-version subdirectories (e.g. `kubernetes/k8s-1.35/`).

To update an existing minor version:

1. Check the [release page](https://github.com/kubernetes/kubernetes/releases).
2. Update `K8S_COMMIT` in `kubernetes/k8s-X.Y/Dockerfile`.
3. Update image tag in `kubernetes/k8s-X.Y/README.md`.
4. Update `BRANCH` and `TAG` files in `kubernetes/k8s-X.Y/`.

To add a new minor version:

1. Create a new directory `kubernetes/k8s-X.Y/`.
2. Create `BRANCH` file containing `X.Y`.
3. Create `TAG` file containing `X.Y.Z.1`.
4. Create `Dockerfile` based on other versions as a reference.
5. Create `README.md` following the same format as other versions.
6. Add the new build job entry to `.github/workflows/main.yaml`.
7. Remove the oldest minor version directory if more than 3 minor versions exist (keep only the latest 3).
