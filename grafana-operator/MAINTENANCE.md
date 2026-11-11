# grafana-operator

1. Check the [release page](https://github.com/grafana/grafana-operator/releases).
2. Check the upstream build procedure (Makefile, Dockerfile, .ko.yaml, etc). At the point of v5.4.1, grafana-operator is built by [ko](https://github.com/ko-build/ko) with its default configuration.
   If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/grafana-operator/tree/vX.Y.Z`
3. Update `VERSION` in `Dockerfile`.
4. Update `TAG`.
5. Update `GRAFANA_OPERATOR_VERSION` in `admission/Makefile`.
6. Follow maintenance instructions for [neco-admission](../admission/MAINTENANCE.md).

> [!Note]
> Grafana Operator's GrafanaDashboard objects are validated by [neco-admission](../admission/MAINTENANCE.md).
> You need to update [neco-admission](../admission/MAINTENANCE.md) at least for its `Makefile`.
