[![Docker Repository on Quay](https://quay.io/repository/cybozu/cke-tools/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/cke-tools)

cke-tools container
===================

This directory provides a Dockerfile to build a Docker container
that runs [cke-tools](https://github.com/cybozu-go/cke-tools).

Usage
-----

### Run rivers: an TCP reverse proxy

For docker:
```console
$ docker run -d --read-only \
    --network host --name cke-tools \
    --entrypoint /usr/local/cke-tools/bin/rivers \
    quay.io/cybozu/cke-tools:0 \
      --listen localhost:6443 \
      --upstreams 10.0.0.100:6443,10.0.0.101:6443,10.0.0.102:6443 
```

For rkt:
```console
$ sudo rkt run \
    --net=host --dns=host \
  docker://quay.io/cybozu/cke-tools:0 \
    --name cke-tools --readonly-rootfs=true \
    --exec /usr/local/cke-tools/bin/rivers \
    -- \
    --listen localhost:6443 \
    --upstreams 10.0.0.100:6443,10.0.0.101:6443,10.0.0.102:6443 
```
