victoriametrics-datasource
==========================

VictoriaMetrics datasource plugin for Grafana

- [victoriametrics-datasource](https://github.com/VictoriaMetrics/victoriametrics-datasource)

This image is intended to be used as an init container.
The entrypoint copies plugin assets to `${GRAFANA_PLUGINS_DIR}` (default:`/var/lib/grafana/plugins`) directory.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/victoriametrics-datasource)
