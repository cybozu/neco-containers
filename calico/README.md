[![Docker Repository on Quay](https://quay.io/repository/cybozu/calico/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/calico)

calico container
======================

This directory provides a Dockerfile to build a Docker container that contains
[calico-node](https://github.com/projectcalico/node) and [calico-typha](https://github.com/projectcalico/typha)
to enable `NetworkPolicy` on Kubernetes cluster.

Usage
-----

### Start `calico`

Run the container

```console
$ docker run -d --read-only --name=calico \
    quay.io/cybozu/calico:3.7.2 start_runit
$ docker run -d --read-only --name=calico --entrypoint="tini --"\
    quay.io/cybozu/calico:3.7.2 calico-typha
```
