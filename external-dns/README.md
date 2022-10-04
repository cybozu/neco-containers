ExternalDNS container
=====================

Build Docker container image for [ExternalDNS][], which synchronizes exposed Kubernetes Services and Ingresses with DNS providers.


Usage
-----

### Start `external-dns`

Run the container

```console
$ docker run -d --read-only --name=external-dns \
    quay.io/cybozu/external-dns:0.12.2.1 \
    --registry=txt --txt-owner-id ... --provider ...
```

[ExternalDNS]: https://github.com/kubernetes-incubator/external-dns/

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/external-dns)
