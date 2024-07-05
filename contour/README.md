Contour container
=================

Build Docker container image for [Contour][], Kubernetes ingress controller using Lyft's Envoy proxy.

Usage
-----

### Start `contour`

Run the container

```console
$ docker run -d --read-only --name=contour \
    ghcr.io/cybozu/contour:1.29.1 serve
```

[Contour]: https://github.com/heptio/contour

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/contour)
