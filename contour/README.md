Contour container
=================

Build Docker container image for [Contour][], Kubernetes ingress controller using Lyft's Envoy proxy.

Usage
-----

### Start `contour`

Run the container

```console
$ docker run -d --read-only --name=contour \
    quay.io/cybozu/contour:latest serve
```

[Contour]: https://github.com/heptio/contour

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/contour)
