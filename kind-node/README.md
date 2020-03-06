[![Docker Repository on Quay](https://quay.io/repository/cybozu/kind-node/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/kind-node)

kind-node container
===================

This directory provides container image `kind-node` and contains customized ptp plugin source code.
This image is used for [kind](https://github.com/kubernetes-sigs/kind) as node container.

Usage
-----

```console
$ kind create cluster --image=quay.io/cybozu/kind-node:1.17
```
