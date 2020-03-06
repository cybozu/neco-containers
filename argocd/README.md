[![Docker Repository on Quay](https://quay.io/repository/cybozu/argocd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/argocd)

# Argo CD container

This directory provides a Dockerfile to build a argocd container
that runs [argoproj/argo-cd](https://github.com/argoproj/argo-cd).

## Usage

### Install `argocd` cli tool to host file system

```console
$ docker run --rm -u root:root \
    --entrypoint /usr/local/argocd/install-tools \
    --mount type=bind,src=DIR,target=/host \
    quay.io/cybozu/argocd:1.3
```

### Deploy argocd-application-controller, argocd-repo-server and argocd-server on k8s

```console
$ kubectl apply -f install.yml
```
