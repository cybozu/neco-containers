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
$ docker run -d --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE \
    --network host --name sabakan \
    quay.io/cybozu/sabakan:0 -etcd-servers http://foo.bar:2379,http://zot.bar:2379
```

### Use client tools

`sabactl` can be used to control sabakan:

```
$ docker exec -it sabakan sabactl -h
```
