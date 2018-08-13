[![Docker Repository on Quay](https://quay.io/repository/cybozu/etcd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/etcd)

etcd-container
==============

[etcd](https://github.com/coreos/etcd) is a distributed reliable key-value
store provided by CoreOS.  This repository provides a Dockerfile which contains
`etcd` server and `etcdctl` for the client usage.

Usage
-----

To launch `etcd` by `docker run`:

    $ docker volume create etcd
    $ docker run -p 2379:2379 -p 2380:2380 --name etcd-1 \
      --mount type=volume,src=etcd,target=/var/lib/etcd \
      quay.io/cybozu/etcd:3.3 \
        --advertise-client-urls http://0.0.0.0:2379 \
        --listen-client-urls http://0.0.0.0:2379

To use `etcdctl`, first install it in a host directory `DIR`:

    $ docker run --rm -u root:root \
      --entrypoint /usr/local/etcd/install-tools \
      --mount type=bind,src=DIR,target=/host \
      quay.io/cybozu/etcd:3.3

Then run `etcdctl` as follows:

    $ DIR/etcdctl get /
