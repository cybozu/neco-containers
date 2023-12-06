dex container
=================

Build Docker container image for [dex][], which is OpenID Connect Identity (OIDC) and OAuth 2.0 Provider with Pluggable Connectors.

Usage
-----

### Start `dex`

Run the container

```console
$ docker run -d --read-only --name=dex \
    ghcr.io/cybozu/dex:2.36
```

[dex]: https://github.com/dexidp/dex

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/dex)
