[![Docker Repository on Quay](https://quay.io/repository/cybozu/alertmanager/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/alertmanager)

alertmanager container
=================

This directory provides a Dockerfile to build a Docker container
that runs [alertmanager](https://github.com/prometheus/alertmanager/).

Usage
-----

### Run the container

For docker:
```console
# create directory to store data
$ sudo mkdir -p /data

$ docker run -d --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE \
    -p 9093:9093 --name alertmanager \
    --mount type=bind,source=/data,target=/data \
    --mount type=bind,source=/config,target=/config \
    quay.io/cybozu/alertmanager:0.15 \
    --config.file=/config/alertmanager.yaml
```

### Use client tools

`amtool` can be used to control alertmanager:

```console
$ docker exec -it alertmanager amtool
```
