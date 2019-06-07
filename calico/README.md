[![Docker Repository on Quay](https://quay.io/repository/cybozu/calico/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/calico)

calico container
================

This directory provides a Dockerfile to build a Docker container that contains
[calico-node](https://github.com/projectcalico/node) and [calico-typha](https://github.com/projectcalico/typha)
to enable `NetworkPolicy` on Kubernetes cluster, and it is not originally included [BIRD][] and [confd][] for dynamic IP routing.

Usage
-----

### Start `calico`

Run the container

```console
# Run as calico-node
$ docker run -d --read-only --name=calico \
    quay.io/cybozu/calico:3.7.2 start_runit

# Run as calico-typha
$ docker run -d --read-only --name=calico --entrypoint="tini --"\
    quay.io/cybozu/calico:3.7.2 calico-typha
```

[BIRD]: https://github.com/projectcalico/bird
[confd]: https://github.com/projectcalico/confd
