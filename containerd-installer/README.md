[![Docker Repository on Quay](https://quay.io/repository/cybozu/containerd-installer/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/containerd-installer)

Containerd installer container
==============================

This directory provides a Dockerfile to build a Docker container
that installs [containerd][] and [cri-tools][].

Usage
-----

### Install `argocd` cli tool to host file system

```console
$ docker run --rm -u root:root \
    --entrypoint /usr/local/containerd-installer/install-tools \
    --mount type=bind,src=DIR,target=/host \
    quay.io/cybozu/containerd-installer:1.2
```

[containerd]: https://github.com/containerd/containerd
[cri-tools]: https://github.com/kubernetes-sigs/cri-tools
