[![Docker Repository on Quay](https://quay.io/repository/cybozu/gorush/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/gorush)

Gorush container
==================

Build Docker container image for [Gorush][], which is a push notification micro server.

Usage
-----

### Run gorush:

```console
$ docker run -d --rm --read-only \
    -p 8088:8088 \
    --name gorush \
    --mount type=bind,source=/home/cybozu/config,target=/config \
    quay.io/cybozu/gorush:1.11.2
```

[Gorush]: https://github.com/appleboy/gorush
