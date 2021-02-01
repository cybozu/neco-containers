Prometheus container
====================

This repository contains Dockerfile for [pushgateway](https://github.com/prometheus/pushgateway).

## Usage

### Run pushgateway:

```console
$ docker run -d --read-only --cap-drop ALL \
    -p 9091:9091 \
    --name pushgateway \
    quay.io/cybozu/pushgateway:1.4
```

## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/prometheus)
