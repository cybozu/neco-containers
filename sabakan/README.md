[![Docker Repository on Quay](https://quay.io/repository/cybozu/sabakan/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/sabakan)

Sabakan container
=================

This directory provides a Dockerfile to build a Docker container
that runs [sabakan](https://github.com/cybozu-go/sabakan).

Usage
-----

### Run the container

For docker:
```
$ docker run -d --read-only --name sabakan \
    -p 67:10067 -p 80:10080 \
    quay.io/cybozu/sabakan:0 -etcd-servers http://foo.bar:2379,http://zot.bar:2379
```

### Use client tools

`sabactl` is an interactive client:

```
$ docker exec -it sabakan sabactl -h
```
