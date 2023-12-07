Prometheus container
====================

This repository contains Dockerfile for [pushgateway](https://github.com/prometheus/pushgateway).

## Usage

### Run pushgateway:

```console
$ docker run -d --read-only --cap-drop ALL \
    -p 9091:9091 \
    --name pushgateway \
    ghcr.io/cybozu/pushgateway:1.4
```

## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/pushgateway)
