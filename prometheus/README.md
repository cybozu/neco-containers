[![Docker Repository on Quay](https://quay.io/repository/cybozu/prometheus/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/prometheus)

Prometheus container
====================

This container contains following applications of Prometheus project.

- [Prometheus](https://github.com/prometheus/prometheus/)
- [Alertmanager](https://github.com/prometheus/alertmanager/)
- [Pushgateway](https://github.com/prometheus/pushgateway/)

Usage
-----

### Run prometheus:

```console
$ docker run -d --read-only \                                    
      -p 9090:9090 \                                               
      --name prometheus \                                          
      --mount type=volume,source=myvolume,target=/data \           
      --mount type=bind,source=/home/cybozu/config,target=/config \
      --entrypoint prometheus \                                    
      quay.io/cybozu/prometheus:2.7.1-1 \                          
      --config.file=/config/prometheus.yaml \                      
      --web.enable-lifecycle \                                     
      --storage.tsdb.path="/data"                                  
```

### Run alertmanager:

```console
# create directory to store data
$ sudo mkdir -p /data

$ docker run -d --read-only --cap-drop ALL --cap-add NET_BIND_SERVICE \
    -p 9093:9093 \
    --name alertmanager \
    --mount type=bind,source=/data,target=/data \
    --mount type=bind,source=/config,target=/config \
    --entrypoint alertmanager \
    quay.io/cybozu/prometheus:2.7.1-1 \
    --config.file=/config/alertmanager.yaml
```

### Run pushgateway:

```console
$ docker run -d --read-only --cap-drop ALL \
    -p 9091:9091 \
    --name pushgateway \
    --mount type=myvolume,source=/data,target=/data \
    --entrypoint pushgateway \
    quay.io/cybozu/prometheus:2.7.1-1 \
    --persistence.file=/data/metrics
```
