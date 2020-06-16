ingress-watcher
===============

`ingress-watcher` is a watching agent for L7 endpoints.

Usage
-----

### Start `ingress-watcher`

`ingress-watcher` exports metrics in one of the following two ways:

1. Run a metrics server and return metrics at `GET /metrics`.
    ```bash
    ingress-watcher push \
    --target-addrs yahoo.com \
    --target-addrs yahoo.com \
    --listen-addr localhost:8080
    ```

2. Push and expose the collected metrics via [Pushgateway](https://github.com/prometheus/pushgateway).
    ```bash
    ingress-watcher push \
    --target-addrs www.google.com \
    --target-addrs yahoo.com \
    --job-name job \
    --push-addr localhost:9091 \
    --push-interval 5s
    ```

The flag values can also be defined with a YAML file.

```yaml
targetAddrs:
- www.google.co.jp
- www.google.com
- foo.bar.baz
watchInterval: 10s

# export
listenAddr: localhost:8080

# for push
pushAddr: localhost:9091
jobName: job
pushInterval: 10s
```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/ingress-watcher)
