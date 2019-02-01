[![Docker Repository on Quay](https://quay.io/repository/cybozu/prometheus/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/prometheus)

Prometheus container
====================

This container contains following applications of Prometheus project.

- [Prometheus](https://github.com/prometheus/prometheus/)
- [Alertmanager](https://github.com/prometheus/alertmanager/)
- [Pushgateway](https://github.com/prometheus/pushgateway/)

Usage
-----

### Install `promtool` to host file system

For docker:
```console
$ docker run --rm -u root:root \
    --entrypoint /usr/local/prometheus/install-tools \
    --mount type=bind,src=DIR,target=/host \
    quay.io/cybozu/prometheus:0
```

