[![CircleCI](https://circleci.com/gh/cybozu/etcd-container.svg?style=svg)](https://circleci.com/gh/cybozu/etcd-container)
[![Docker Repository on Quay](https://quay.io/repository/cybozu/etcd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/etcd)

etcd-container
==============

[etcd](https://github.com/coreos/etcd) is a distributed reliable key-value
store provided by CoreOS.  This repository provides a Dockerfile which contains
`etcd` server and `etcdctl` for the client usage.

Usage
-----

To launch `etcd` by `docker run`:

    $ docker run -p 2379:2379 -p 2380:2380 --name etcd-1 quay.io/cybozu/etcd:3.2 \
        --name etcd-1 \
        --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379

To use `etcdctl` by `docker run`:

    $ docker run --rm -it --entrypoint etcdctl etcd:3.2 --endpoints ${ETCD_ENDPOINTS} get /

Note that `etcdctl` runs also in the container.  If `--endpoints` is not set,
`etcdctl` try to connects `localhost` in the container.

Lisence
-------

MIT
