[![Docker Repository on Quay](https://quay.io/repository/cybozu/sabakan/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/sabakan)

Sabakan container
=================

This directory provides a Dockerfile to build a Docker container
that runs [sabakan](https://github.com/cybozu-go/sabakan).

Usage
-----

### Run the container

For docker:
```console
# create directory to store OS images
$ sudo mkdir -p /var/lib/sabakan

# -advertise-url is the canonical URL of this sabakan.
$ docker run -d --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE \
    --network host --name sabakan \
    --mount type=bind,source=/var/lib/sabakan,target=/var/lib/sabakan \
    quay.io/cybozu/sabakan:0 \
    -etcd-servers http://foo.bar:2379,http://zot.bar:2379 \
    -advertise-url http://12.34.56.78:10080
```

For rkt:
```console
# create directory to store OS images
$ sudo mkdir -p /var/lib/sabakan

$ sudo rkt run \
    --volume data,kind=host,source=/var/lib/sabakan \
    --net=host --dns=host \
  docker://quay.io/cybozu/sabakan:0 \
    --name sabakan --readonly-rootfs=true \
    --caps-retain=CAP_NET_BIND_SERVICE \
    --mount volume=data,target=/var/lib/sabakan \
    -- \
    -etcd-servers http://foo.bar:2379,http://zot.bar:2379 \
    -advertise-url http://12.34.56.78:10080
```

### Use client tools

`sabactl` can be used to control sabakan:

```console
$ docker exec -it sabakan sabactl -h
```
