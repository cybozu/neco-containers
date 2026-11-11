# victoriametrics

![Regular Update](../regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/VictoriaMetrics/releases).
2. Check upstream `Makefile` and `Dockerfile`, and update our `Dockerfile` if needed.
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/Makefile`
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/app/*/Makefile`
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/app/*/deployment/Dockerfile`
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z-cluster/Makefile`
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z-cluster/app/*/Makefile`
   - `https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z-cluster/app/*/deployment/Dockerfile`
3. Update `VICTORIAMETRICS_SINGLE_VERSION` and `VICTORIAMETRICS_CLUSTER_VERSION` in `Dockerfile`.
4. Update `TAG` file.
