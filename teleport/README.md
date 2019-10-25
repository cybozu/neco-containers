[![Docker Repository on Quay](https://quay.io/repository/cybozu/teleport/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/teleport)

Teleport container
==================

Build Docker container image for [Teleport][], which  is a modern security gateway for remotely accessing.


Usage
-----

### Start `teleport`

Run the container

```console
$ docker run -p 3022:3022 -p 3023:3023 -p 3024:3024 -p 3025:3025 -p 3026:3026 -p 3080:3080 \
    -d --read-only --name=teleport \
    quay.io/cybozu/teleport:4.1.3 \
    start ...
```

Get teleport binary

```console
$ docker run  --rm -u root:root \
    --entrypoint /usr/local/teleport/install-tools \
    --mount type=bind,source=DIR,target=/host \
    quay.io/cybozu/teleport:4.1.3
```

[Teleport]: https://github.com/gravitational/teleport
