[![Docker Repository on Quay](https://quay.io/repository/cybozu/pause/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/pause)

pause container
===============

[pause container](https://github.com/kubernetes/kubernetes/tree/master/build/pause) works as the parent of all other containers in a pod.

Usage
-----

Specify the image name for kubelet with `--pod-infra-container-image` option.
