[![Docker Repository on Quay](https://quay.io/repository/cybozu/teleport/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/teleport)

Teleport container
==================

Build Docker container image for [Teleport][], which  is a modern security gateway for remotely accessing.


Usage
-----

### Start `teleport`

Run the container

```console
$ docker run -d --read-only --name=teleport \
    quay.io/cybozu/teleport:4.0.2 \
    start ...
```

[Teleport]: https://github.com/gravitational/teleport
