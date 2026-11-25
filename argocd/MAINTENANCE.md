# argocd

ArgoCD images are split into per-minor-version directories (`argocd/argocd-X.Y/`).
Each `argocd-X.Y/` directory contains `argocd` and the images it depends on: `dex`, `haproxy`, `redis` and `redis_exporter`.
The procedures below apply to every `argocd-X.Y/` directory.

## argocd

1. Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.
2. Check `hack/tool-versions.sh` for the tools versions.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/hack/tool-versions.sh`
3. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/Dockerfile`
4. Update version variables in `Dockerfile` to the versions in `hack/tool-versions.sh`.
    - Update `ARGOCD_VERSION`, `KUSTOMIZE_VERSION`, `HELM_VERSION` and `GIT_LFS_VERSION`.
    - Update the `--checksum` of each download. Upstream keeps the checksums of the tools in `hack/installers/checksums/`.
      Compare them with the ones published by each project.
        - [Kustomize](https://github.com/kubernetes-sigs/kustomize/releases)
        - [Helm](https://github.com/helm/helm/releases)
        - [Git LFS](https://github.com/git-lfs/git-lfs/releases)
    - Calculate the checksum of the Argo CD source archive with `sha256sum`.
5. Update `BRANCH` and `TAG` files.

> [!Note]
> ArgoCD depends on dex,Redis,HAProxy.
> So browse the following manifests and update [dex](#dex),[redis](#redis),[haproxy](#haproxy) images next.

- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

## dex

> [!Note]
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`

1. Check the [release page](https://github.com/dexidp/dex/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/dexidp/dex/blob/vX.Y.Z/Dockerfile`
3. Update `DEX_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## haproxy

> [!Note]
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

1. Check the release notes in the [official site](https://www.haproxy.org/).
   - v2.6.x: <https://github.com/docker-library/haproxy/blob/master/2.6/Dockerfile>
2. Update `HAPROXY_SHA256` in `Dockerfile`, SHA256 hash in <http://www.haproxy.org/download>
3. Update `BRANCH` and `TAG` files.

## redis

> [!Note]
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`

1. Check the release notes in the [official site](https://redis.io/).
2. Check the Dockerfile in docker-library. If there are any updates, update our `Dockerfile`.
   - v7.0.x: <https://github.com/docker-library/redis/blob/master/7.0/Dockerfile>
3. Update `REDIS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
5. Update `README.md`, `COPYING` and the container LABEL in terms of license if necessary.

## redis_exporter

1. Check the [release page](https://github.com/oliver006/redis_exporter/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
3. Update `REDIS_EXPORTER_VERSION` and checksum of archive in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
