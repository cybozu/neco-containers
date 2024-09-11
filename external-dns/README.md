ExternalDNS container
=====================

Build Docker container image for [ExternalDNS][], which synchronizes exposed Kubernetes Services and Ingresses with DNS providers.


Usage
-----

### Start `external-dns`

Run the container

```console
$ docker run -d --read-only --name=external-dns \
    ghcr.io/cybozu/external-dns:0.15.0.1 \
    --registry=txt --txt-owner-id ... --provider ...
```

[ExternalDNS]: https://github.com/kubernetes-incubator/external-dns/

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/external-dns)
