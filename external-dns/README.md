[![Docker Repository on Quay](https://quay.io/repository/cybozu/external-dns/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/external-dns)

ExternalDNS container
=====================

Build Docker container image for [ExternalDNS][], which synchronizes exposed Kubernetes Services and Ingresses with DNS providers.


Usage
-----

### Start `external-dns`

Run the container

```console
$ docker run -d --read-only --name=external-dns \
    quay.io/cybozu/external-dns:0.5.13 \
    --registry=txt --txt-owner-id ... --provider ...
```

[ExternalDNS]: https://github.com/kubernetes-incubator/external-dns/
