[![Docker Repository on Quay](https://quay.io/repository/cybozu/node-exporter/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/node-exporter)

Node exporter container
=======================

[Node exporter](https://github.com/prometheus/node_exporter) is
Prometheus exporter for hardware and OS metrics exposed by *NIX kernels.

Usage
------

To launch `node_exporter` by `docker run`:

```console
$ docker run -d \
    --net="host" \
    --pid="host" \
    -v "/:/host:ro,rslave" \
    quay.io/cybozu/node-exporter:0.17.0-1 \
    --path.rootfs /host
```
