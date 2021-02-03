Prometheus container
====================

This repository contains Dockerfile for [Prometheus](https://prometheus.io/).

## Usage

```console
$ docker run -d --read-only \                                    
      -p 9090:9090 \                                               
      --name prometheus \                                          
      --mount type=volume,source=myvolume,target=/data \           
      --mount type=bind,source=/home/cybozu/config,target=/etc/prometheus \
      quay.io/cybozu/prometheus:2.24
```

## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/prometheus)
