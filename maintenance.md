# How to maintain neco-containers <!-- omit in toc -->

This document describes the procedure for updating each container image.

Besides on-demand container updates, we have two regular renewal operations: `Kubernetes Update` and `Regular Update`.
The target container of these operations have the following badges, so check before the operations.

In case of components whose Go source code are in neco-containers, all dependent Go modules should be updated if there is no special reason. Kubernetes-related modules such as client-go may be newer than the Kubernetes to be updated. For example, it is acceptable that client-go is v0.30 and Kubernetes is v1.29.

1. Kubernetes Update (![Kubernetes Update](./kubernetes_update.svg))
   - Upgrade of Kubernetes. Besides the related components of Kubernetes, update the containers managed by [CKE](https://github.com/cybozu-go/cke/) and some go modules.
2. Regular Update (![Regular Update](./regular_update.svg))
   - Update in every quarter. Keeping up with the upstream version and updating the ubuntu base image.
3. CSA Update  (![CSA Update](./csa_update.svg))
   - Update by CSA team.
4. No Need Update (![No Need Update](./no_need_update.svg))
   - Used as a PoC, so regular updates are not required.

---

- [alertmanager](#alertmanager)
- [alloy](#alloy)
- [blackbox\_exporter](#blackbox_exporter)
- [fluent-bit](#fluent-bit)
- [grafana](#grafana)
- [grafana-operator](#grafana-operator)
- [loki](#loki)
- [memcached](#memcached)
- [memcached\_exporter](#memcached_exporter)
- [opentelemetry-collector](#opentelemetry-collector)
- [prometheus-config-reloader](#prometheus-config-reloader)
- [pushgateway](#pushgateway)
- [pyroscope](#pyroscope)
- [tempo](#tempo)
- [victoriametrics](#victoriametrics)
- [victoriametrics-operator](#victoriametrics-operator)

---

## alertmanager

![Regular Update](./regular_update.svg)

1. Check the release page.
   - <https://github.com/prometheus/alertmanager/releases>
2. Check the upstream Makefile and Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Makefile`
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/prometheus/alertmanager/blob/vX.Y.Z/ui/Dockerfile` (check Node.js version)
3. Update version variables in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## alloy

![Regular Update](./regular_update.svg)

Alloy contains two versions, one for alloy and the other for libsystemd.
The libsystemd version should be the same with the one running on [the stable Flatcar OS](https://www.flatcar.org/releases).

1. Check the release page.
   - <https://github.com/grafana/alloy/releases>
2. Check the upstream Dockerfile and Makefile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/.github/workflows/build.yml`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/tools/build-image/Dockerfile`
3. Update `ALLOY_VERSION` in `Dockerfile`.
4. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
5. Update `TAG` file.

## blackbox_exporter

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/prometheus/blackbox_exporter/releases).
2. Update `BLACKBOX_EXPORTER_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` files.

## fluent-bit

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/fluent/fluent-bit/releases).
2. Update `FLUENT_BIT_VERSION` in `Dockerfile`.
3. Update `TAG`.

## grafana

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/grafana/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/grafana/blob/vX.Y.Z/Dockerfile`
   - Check `JS_IMAGE` in the Dockerfile
3. Update `GRAFANA_VERSION` in `Dockerfile`.
4. Update installation of Node.js in `Dockerfile` according to `JS_IMAGE` if necessary.
5. Update `TAG`.

## grafana-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/grafana-operator/releases).
2. Check the upstream build procedure (Makefile, Dockerfile, .ko.yaml, etc). At the point of v5.4.1, grafana-operator is built by [ko](https://github.com/ko-build/ko) with its default configuration.
   If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/grafana-operator/tree/vX.Y.Z`
3. Update `VERSION` in `Dockerfile`.
4. Update `TAG`.
5. Update `GRAFANA_OPERATOR_VERSION` in `admission/Makefile`.
6. Follow maintenance instructions for [neco-admission](./admission/MAINTENANCE.md).

> [!Note]
> Grafana Operator's GrafanaDashboard objects are validated by [neco-admission](./admission/MAINTENANCE.md).
> You need to update [neco-admission](./admission/MAINTENANCE.md) at least for its `Makefile`.

## loki

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/loki/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/loki/blob/vX.Y.Z/cmd/loki/Dockerfile`
3. Update `LOKI_VERSION` in `Dockerfile`.
4. Update `TAG` file.

## memcached

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/memcached/memcached/wiki/ReleaseNotes).
2. Update `MEMCACHED_VERSION` in `Dockerfile`.
3. Update `TAG` file.

## memcached_exporter

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/prometheus/memcached_exporter/releases).
2. Update `MEMCACHED_EXPORTER_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` file.

## opentelemetry-collector

![Regular Update](./regular_update.svg)

opentelemetry-collector container consists of three repositories: opentelemetry-collector, opentelemetry-collector-contrib and opentelemetry-collector-releases

1. Check the release pages [main](https://github.com/open-telemetry/opentelemetry-collector/releases) [contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases) [release](https://github.com/open-telemetry/opentelemetry-collector-releases/releases).
2. Check the upstream Dockerfile and builder manifest. If there are any updates, update our `Dockerfile`.
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/Dockerfile`
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/manifest.yaml`
3. Update `OTELCOL_VERSION` and/or `OTELCOL_CONTRIB_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## prometheus-config-reloader

![Regular Update](./regular_update.svg)

This is a part of [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator/).
This is used as a sidecar of Alloy Pod now.

1. Check the latest release of `prometheus-operator`
2. Update version variable in `Dockerfile`.
3. Update `TAG` file.

## pushgateway

![Regular Update](./regular_update.svg)

1. Check the release page.
   - <https://github.com/prometheus/pushgateway/releases>
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/prometheus/pushgateway/blob/vX.Y.Z/Dockerfile`
3. Update version variables in `Dockerfile`.
4. Update `TAG` file.

## pyroscope

![Regular Update](./regular_update.svg)

1. Check the pyroscope [release page](https://github.com/grafana/pyroscope/releases).
2. Check the pyroscope [Dockerfile](https://github.com/grafana/pyroscope/blob/main/cmd/pyroscope/Dockerfile) of the latest release version.
3. Update `TAG` file.

## tempo

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/tempo/releases).
2. Check the upstream `Makefile` and `cmd/tempo//Dockerfile`. If they have been updated significantly, update our `Dockerfile`.
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/cmd/tempo/Dockerfile`
3. Update `TEMPO_VERSION` in `Dockerfile`.
4. Update `TAG` file.

## victoriametrics

![Regular Update](./regular_update.svg)

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

## victoriametrics-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/operator/releases).
2. Check upstream Makefile and Dockerfile, and update our Dockerfile if needed.
3. Update `VICTORIAMETRICS_OPERATOR_VERSION` in `Dockerfile`.
4. Update `TAG` file.
