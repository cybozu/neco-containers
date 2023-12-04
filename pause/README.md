pause container
===============

[pause container](https://github.com/kubernetes/kubernetes/tree/master/build/pause) works as the parent of all other containers in a pod.

Usage
-----

Specify the image name for kubelet with `--pod-infra-container-image` option.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/pause)
