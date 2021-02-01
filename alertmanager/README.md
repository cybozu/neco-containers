Prometheus Alertmanager container
=================================

This repository contains Dockerfile for [alertmanager](https://github.com/prometheus/alertmanager/).

## Usage

```console
# create directory to store data
$ sudo mkdir -p /data

$ docker run -d --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE \
    -p 9093:9093 \
    --name alertmanager \
    --mount type=bind,source=/data,target=/data \
    --mount type=bind,source=/config,target=/config \
    --entrypoint alertmanager \
    quay.io/cybozu/alertmanager:0.21.0.2 \
    --config.file=/config/alertmanager.yaml
```

## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/alertmanager)
