[![Docker Repository on Quay](https://quay.io/repository/cybozu/contour/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/contour)

Contour container
=================

Build Docker container image for [Contour][], Kubernetes ingress controller using Lyft's Envoy proxy.

Usage
-----

### Start `contour`

Run the container

```console
$ docker run -d --read-only --name=contour \
    quay.io/cybozu/contour:1.0 serve
```

[Contour]: https://github.com/heptio/contour
