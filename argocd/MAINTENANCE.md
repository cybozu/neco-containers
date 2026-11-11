# argocd

![Regular Update](../regular_update.svg)

1. Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.
2. Check `hack/tool-versions.sh` for the tools versions.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/hack/tool-versions.sh`
3. Update tool versions in `Dockerfile`.
    - [Kustomize](https://github.com/kubernetes-sigs/kustomize/releases)
    - [Helm](https://github.com/helm/helm/releases)
4. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/Dockerfile`
5. Update version variables in `Dockerfile`.
    - Update `ARGOCD_VERSION`, `KUSTOMIZE_VERSION` and `HELM_VERSION`.
6. Update `BRANCH` and `TAG` files.
7. Update `ARGOCD_VERSION` in `admission/Makefile`.
8. Follow maintenance instructions for [neco-admission](../admission/MAINTENANCE.md).

> [!Note]
> ArgoCD depends on dex,Redis,HAProxy.
> So browse the following manifests and update [dex](../maintenance.md#dex),[redis](../maintenance.md#redis),[haproxy](../maintenance.md#haproxy) images next.

- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

> [!Note]
> ArgoCD's Application objects are validated by [neco-admission](../admission/MAINTENANCE.md).
> You need to update [neco-admission](../admission/MAINTENANCE.md) at least for its `Makefile`.
