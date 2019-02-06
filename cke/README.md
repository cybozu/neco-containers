[![Docker Repository on Quay](https://quay.io/repository/cybozu/cke/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/cke)

CKE container
=============

This directory provides a Dockerfile to build a Docker container
that runs [CKE](https://github.com/cybozu-go/cke).

Usage
-----

### Run CKE

For docker:
```console
$ docker run -d --read-only \
    --network host --name cke \
    quay.io/cybozu/cke:1.13 [options...]
```

For rkt:
```console
$ sudo rkt run \
    --net=host --dns=host \
  docker://quay.io/cybozu/cke:1.13 \
    --name cke --readonly-rootfs=true \
    -- [options...]
```

### Install ckecli to host file system

For docker:
```console
$ docker run --rm -u root:root \
    --entrypoint /usr/local/cke/install-tools \
    --mount type=bind,src=DIR,target=/host \
    quay.io/cybozu/cke:1.13
```

For rkt:
```console
$ sudo rkt run \
    --volume host,kind=host,source=DIR \
    --mount volume=host,target=/host \
    --exec /usr/local/cke/install-tools \
    quay.io/cybozu/cke:1.13
```
