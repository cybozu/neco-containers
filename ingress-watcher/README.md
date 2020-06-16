ingress-watcher
===============

`ingress-watcher` is a watching agent for L7 endpoints.

Usage
-----

### Start `ingress-watcher`

`ingress-watcher` exports metrics in one of the following two ways:

1. Run a metrics server and return metrics at `GET /metrics`.
    ```bash
    ingress-watcher export \
    --target-addrs example.com \
    --target-addrs example.org \
    --listen-addr localhost:8080 \
    --watch-interval 10s
    ```

2. Push and expose the collected metrics via [Pushgateway](https://github.com/prometheus/pushgateway).
    ```bash
    ingress-watcher push \
    --target-addrs example.com \
    --target-addrs example.org \
    --push-addr localhost:9091 \
    --watch-interval 10s \
    --job-name job \
    --push-interval 5s
    ```

The flag values can also be defined with a YAML file with the flag `--config <filename>`. Flag values are overwritten by this YAML file.

```yaml
targetAddrs:
- www.google.co.jp
- www.google.com
- foo.bar.baz
watchInterval: 10s

# for export
listenAddr: localhost:8080

# for push
pushAddr: localhost:9091
jobName: job
pushInterval: 10s
```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/ingress-watcher)
