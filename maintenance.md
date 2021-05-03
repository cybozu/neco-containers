How to maintain neco-containers
===============================

This document describes the procedure for updating each container image.

Besides on-demand container updates, we have two regular renewal operations: `Kubernetes Update` and `Regular Update`.
The target container of these operations have the following badges, so check before the operations.

1 Kubernetes Update (![Kubernetes Update](./kubernetes_update.svg))
   - Upgrade of Kubernetes. Besides the related components of Kubernetes,  update the containers managed by [CKE](https://github.com/cybozu-go/cke/) and some go modules.

2 Regular Update (![Regular Update](./regular_update.svg))
   - Update in every quarter. Keeping up with the upstream version and updating the ubuntu base image.

---

## admission (neco-admission)

![Kubernetes Update](./kubernetes_update.svg)

1. Update version variables in `Makefile`.
2. Update go modules.
3. Modify the code to match the new CRDs if CRDs are changed.
   - The code which depended on the CRDs are in the [hook](https://github.com/cybozu/neco-containers/tree/main/admission/hooks) directory.
   - And let's use `Unstructured` instead of use golang library. Take a look at [this PR](https://github.com/cybozu/neco-containers/pull/339/files).
4. Generate code and manifests.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/admission
   $ make setup
   $ make generate manifests
   # Commit, if there are any updated files.
   ```
5. Confirm build and test are green.
   ```bash
   $ make build test
   ```
6. Update `TAG` file.

## alertmanager

![Regular Update](./regular_update.svg)

1. Check the release page.
   - https://github.com/prometheus/alertmanager/releases
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Dockerfile
3. Update version variables in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## argocd

![Kubernetes Update](./kubernetes_update.svg) 

(only for `kubectl` inside Docker image)

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.
2. Check [`hack/tool-versions.sh`](https://github.com/argoproj/argo-cd/blob/master/hack/tool-versions.sh) for the tools versions, especially the version of `packr`.
3. Update tool versions in `Dockerfile`
    - [Kustomize](https://github.com/kubernetes-sigs/kustomize/releases)
    - [Helm](https://github.com/helm/helm/releases)
4. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - https://github.com/argoproj/argo-cd/blob/vX.Y.Z/Dockerfile
5. Update version variables in `Dockerfile`.
    - In the regular update: Update `ARGOCD_VERSION`, `KUSTOMIZE_VERSION`, `HELM_VERSION` and `PACKR_VERSION`.
    - In the Kubernetes update: Update `KUSTOMIZE_VERSION`.
6. Update `BRANCH` and `TAG` files.

## bird

![Regular Update](./regular_update.svg)

1. Check the [releases page](https://bird.network.cz/?download) in the official website.
2. Update `BIRD_VERSION` variable in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## bmc-reverse-proxy

![Kubernetes Update](./kubernetes_update.svg)

1. Update go modules.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/bmc-reverse-proxy
   $ K8SLIB_VERSION=X.Y.Z # e.g. K8SLIB_VERSION=0.18.9
   $ go get k8s.io/apimachinery@v$K8SLIB_VERSION k8s.io/client-go@v$K8SLIB_VERSION
   $ go mod tidy
   ```
2. Confirm test are green.
   ```bash
   $ make test
   ```
3. Update image tag in `bmc-reverse-proxy.yaml`.
4. Update `TAG` file.

## calico

![Regular Update](./regular_update.svg)

1. Check [the release notes](https://docs.projectcalico.org/release-notes/).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/projectcalico/node/blob/vX.Y.Z/Dockerfile.amd64
   - https://github.com/projectcalico/typha/blob/vX.Y.Z/docker-image/Dockerfile.amd64
3. Update version variables (`CALICO_VERSION` and `TINI_VERSION`) in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## ceph

1. Check the [release page](https://docs.ceph.com/en/latest/releases/).
2. Update the `version` argument on the `build-ceph` job in the CircleCI `main` workflow.
3. Update `BRANCH` and `TAG` files.

***NOTE:*** The rook image is based on the ceph image. So upgrade the rook image next.

TODO: Please add how to maintain Dockerfile. I don't know the URL of the upstream Dockerfile.

## cert-manager

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/jetstack/cert-manager/releases) for changes.
2. Update the `version` argument on the `build-cert-manager` job in the CircleCI `main` workflow.
   - If the build fails, please check the Bazel version which is defined as `BAZEL_VERSION` in `build-cert-manager` job.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## chrony

![Regular Update](./regular_update.svg)

1. Check the [release note](https://chrony.tuxfamily.org/news.html).
2. Update `CHRONY_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## configmap-reload

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/jimmidyson/configmap-reload/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/jimmidyson/configmap-reload/blob/vX.Y.Z/Dockerfile
3. Update `CONFIGMAP_RELOAD_VERSION` in `Dockerfile`
4. Update `src/CHANGELOG.md`.
5. Update `BRANCH` and `TAG` files.

## consul

1. Check the [release page](https://github.com/hashicorp/consul/releases).
2. Update `CONSUL_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` file.

## contour

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/projectcontour/contour/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/projectcontour/contour/blob/vX.Y.Z/Dockerfile
3. Update `CONTOUR_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## coredns

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/coredns/coredns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/coredns/coredns/blob/vX.Y.Z/Dockerfile
3. Update `COREDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## dex

![Regular Update](./regular_update.svg)

***NOTE:*** This image is used by ArgoCD. So browse the following manifest and check the required version.
- https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml

1. Check the [release page](https://github.com/dexidp/dex/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/dexidp/dex/blob/vX.Y.Z/Dockerfile
3. Update `DEX_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## dnsmasq

![Regular Update](./regular_update.svg)

1. Check the http://www.thekelleys.org.uk/dnsmasq/ and find the latest release.
2. Update `DNSMASQ_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## envoy

![Regular Update](./regular_update.svg)

***NOTE:*** Envoy is managed by Contour so update to the supported version. See the below.
- [Contour Compatibility Matrix](https://projectcontour.io/resources/compatibility-matrix/)

1. Check the [release page](https://github.com/envoyproxy/envoy/releases).
2. Update the `version` argument on the `build-envoy` job in the CircleCI `main` workflow.
3. Update `BAZEL_VERSION` in `build-envoy` job. The required version is written in the following file.
   - https://github.com/envoyproxy/envoy/blob/vX.Y.Z/.bazelversion
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## etcd

![Kubernetes Update](./kubernetes_update.svg)

***NOTE:*** Upgrading to etcd 3.4+ will require modifications to CKE, so it should be done separately.

1. Check the [release page](https://github.com/etcd-io/etcd/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/etcd-io/etcd/blob/vX.Y.Z/Dockerfile-release
3. Update `ETCD_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## external-dns

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/external-dns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/kubernetes-sigs/external-dns/blob/vX.Y.Z/Dockerfile
3. Update `EXTERNALDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `TAG` file.

## fluent-bit

![Regular Update](./regular_update.svg)

Fluent Bit contains two versions, one for Fluent Bit and the other for libsystemd.
The libsystemd version should be the same with the one running on [the stable Flatcar OS](https://kinvolk.io/flatcar-container-linux/releases/).

1. Check the [release page](https://github.com/fluent/fluent-bit/releases).
2. Update `FLUENT_BIT_VERSION` in `Dockerfile`.
3. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
4. Update `BRANCH` and `TAG`.

## golang / golang-bionic

![Regular Update](./regular_update.svg)

1. Check the [release history](https://golang.org/doc/devel/release.html).
2. Update `GO_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG`.

## gorush

Ignore!!!

## grafana

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/grafana/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/grafana/grafana/blob/vX.Y.Z/Dockerfile
   - Check `NODEVERSION` in https://github.com/grafana/grafana/blob/vX.Y.Z/scripts/build/ci-build/Dockerfile
3. Update `GRAFANA_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## grafana_plugins_init

![Regular Update](./regular_update.svg)

grafana_plugins_init does not create [release](https://github.com/integr8ly/grafana_plugins_init/releases). Use the revision which the operator uses.

1. Check `PluginsInitContainerTag` in [pkg/controller/config/controller_config.go of grafana-operator](https://github.com/integr8ly/grafana-operator/blob/master/pkg/controller/config/controller_config.go).
2. Check [the commit history](https://github.com/integr8ly/grafana_plugins_init/commits/master) and identify which revision correspond to `PluginsInitContainerTag`.
3. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/integr8ly/grafana_plugins_init/blob/xxxxxxxx/Dockerfile
4. Update `REVISION` in `Dockerfile`.
5. Update `BRANCH` and `TAG`.

## grafana-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/integr8ly/grafana-operator/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/integr8ly/grafana-operator/blob/vX.Y.Z/build/Dockerfile
3. Update `VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG`.

## heartbeat

![Regular Update](./regular_update.svg)

Only the base image should be updated.

1. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

## kube-metrics-adapter

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/zalando-incubator/kube-metrics-adapter/releases).
2. Update `KMA_VERSION` in `Dockerfile`.
3. Update `TAG` file.

## kube-state-metrics

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes/kube-state-metrics/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/kubernetes/kube-state-metrics/blob/vX.Y.Z/Dockerfile
3. Update `KUBE_STATE_METRICS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## kubernetes

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes/kubernetes/releases).
2. Update `K8S_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## local-pv-provisioner

![Kubernetes Update](./kubernetes_update.svg)

1. Update version variables in `Makefile`.
2. Update go modules.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/local-pv-provisioner
   $ K8SLIB_VERSION=X.Y.Z # e.g. K8SLIB_VERSION=0.18.9
   $ go get k8s.io/api@v$K8SLIB_VERSION k8s.io/apimachinery@v$K8SLIB_VERSION k8s.io/client-go@v$K8SLIB_VERSION
   $ go get sigs.k8s.io/controller-runtime@v<CTRL_VERSION>
   $ go mod tidy
   ```
3. Generate code and manifests.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/local-pv-provisioner
   $ make setup
   $ make generate manifests
   # Commit, if there are any updated files.
   ```
4. Confirm build and test are green.
   ```bash
   $ make build test
   ```
5. Update image tag in `local-pv-provisioner.yaml`.
6. Update `TAG` file.

## loki

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/loki/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/grafana/loki/blob/vX.Y.Z/cmd/loki/Dockerfile
3. Update `LOKI_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## machines-endpoints

![Kubernetes Update](./kubernetes_update.svg)

1. Update version variables in `Makefile`.
2. Update go modules.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/machines-endpoints
   $ K8SLIB_VERSION=X.Y.Z # e.g. K8SLIB_VERSION=0.18.9
   $ go get k8s.io/api@v$K8SLIB_VERSION k8s.io/apimachinery@v$K8SLIB_VERSION k8s.io/client-go@v$K8SLIB_VERSION
   $ go mod tidy
   ```
3. Confirm test is green.
   ```bash
   $ make test
   ```
4. Update image tag in `machines-endpoints.yaml`.
5. Update `TAG` file.


## memcached

1. Check the [release page](https://github.com/memcached/memcached/wiki/ReleaseNotes).
2. Update `MEMCACHED_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` file.

## memcached_exporter

1. Check the [release page](https://github.com/prometheus/memcached_exporter/releases).
2. Update `MEMCACHED_EXPORTER_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` file.

## metallb

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/metallb/metallb/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/metallb/metallb/blob/vX.Y.Z/controller/Dockerfile
   - https://github.com/metallb/metallb/blob/vX.Y.Z/speaker/Dockerfile
3. Update `METALLB_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## metrics-server

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/metrics-server/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/kubernetes-sigs/metrics-server/blob/vX.Y.Z/Dockerfile
3. Update `METRICS_SERVER_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## moco-mysql

![Regular Update](./regular_update.svg)

moco-mysql is used for [MOCO](https://github.com/cybozu-go/moco).
The MySQL versions are the ones supported by MOCO. So the versions need not update usually.
In the regular update, only update the ubuntu base image.

1. Update all `TAG` files in sub directories.

### moco-backup

`moco-mysql/moco-backup` directory contains a container image used for MOCO's backup/restore feature.
It should include `mysql` and `mysqlbinlog` binaries from the latest MySQL and MySQL shell of the
same version.

The image tag in `TAG` file should therefore be named after the latest MySQL.  For instance,
if the latest MySQL version is 8.0.24, the TAG should be `8.0.24.x` where x is an integer >= 1.

The URL of the MySQL shell debian package can be found in https://dev.mysql.com/downloads/shell/ .

1. Choose "Ubuntu Linux"
2. Choose `mysql-shell_*ubuntu*_amd64.deb` (not a `dbgsym` image) and click "Download" button.
3. Copy the URL from the link whose text reads `No thanks, just start my download.`.
4. Update `MYSQLSH_VERSION` in `moco-mysql/moco-backup/Dockerfile`.

When the latest MySQL version changes, edit `.circleci/config.yml` and update the required job.

```yaml
      - build:
          name: build-moco-backup
          container-image: moco-backup
          attach: true
          dir: moco-mysql/moco-backup
          requires:
            # Update this
            - build-moco-mysql-8024
```

## mysqld_exporter

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/prometheus/mysqld_exporter/releases).
2. Update `MYSQLD_EXPORTER_VERSION` in `Dockerfile`.
3. Update `TAG` file.

## pause

![Kubernetes Update](./kubernetes_update.svg)

1. Check the changelog.
   - https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/CHANGELOG.md
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/Dockerfile
3. Update `PAUSE_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## prometheus-adapter

![Regular Update](./regular_update.svg)

1. Check the release page.
   - https://github.com/kubernetes-sigs/prometheus-adapter/releases
2. Update version variables in `Dockerfile`.
3. Update `TAG` file.

## prometheus-config-reloader

![Regular Update](./regular_update.svg)

This is a part of [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator/).
This is used by victoria-metrics operator too.

1. Check the latest release of `prometheus-operator`
2. Update version variable in `Dockerfile`.
3. Update `BRANCH` and `TAG` files.

## promtail

![Regular Update](./regular_update.svg)

Promtail contains two versions, one for promtail and the other for libsystemd.
The promtail version should be the same with [loki](##loki).
The libsystemd version should be the same with the one running on [the stable Flatcar OS](https://kinvolk.io/flatcar-container-linux/releases/).

1. Update `LOKI_VERSION` in `Dockerfile`.
2. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
3. Update `TAG` file.

## pushgateway

![Regular Update](./regular_update.svg)

1. Check the release page.
   - https://github.com/prometheus/pushgateway/releases
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/prometheus/pushgateway/blob/vX.Y.Z/Dockerfile
3. Update version variables in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## redis

![Regular Update](./regular_update.svg)

***NOTE:*** This image is used by ArgoCD. So browse the following manifest and check the required version.
- https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml

1. Check the release notes in the [official site](https://redis.io/).
2. Check the Dockerfile in docker-library. If there are any updates, update our `Dockerfile`.
   - v6.0: https://github.com/docker-library/redis/blob/master/6.0/Dockerfile
3. Update `REDIS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## registry

![Regular Update](./regular_update.svg)

1. Check the release notes in the [release page](https://github.com/docker/distribution/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/docker/distribution/blob/master/Dockerfile
3. Update `REGISTRY_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## rook

***NOTE:*** The rook image is based on the ceph image. So upgrade the ceph image first.

1. Check the [release page](https://github.com/rook/rook/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Dockerfile
3. Check the `TINI_VERSION` in the following Makefile.
   - https://github.com/rook/rook/blob/vX.Y.Z/images/Makefile
4. Update `ROOK_VERSION` and `TINI_VERSION` in `Dockerfile`.
5. Update ceph image tag in `Dockerfile`.
6. Update `BRANCH` and `TAG` files.

## sealed-secrets

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/bitnami-labs/sealed-secrets/releases).
2. Check the upstream Dockerfile and compare with ours especially on the runtime stage. If there are any updates, update our `Dockerfile`.
    - https://github.com/bitnami-labs/sealed-secrets/blob/vX.Y.Z/docker/Dockerfile
3. Update `SEALED_SECRETS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## serf

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/serf/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - https://github.com/hashicorp/serf/blob/vX.Y.Z/scripts/serf-builder/Dockerfile
3. Update `SERF_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## squid

![Regular Update](./regular_update.svg)

1. Check the latest **stable** version at http://www.squid-cache.org/Versions/
2. Check release notes if a new version is released.
    - e.g., http://www.squid-cache.org/Versions/v4/squid-4.13-RELEASENOTES.html
3. Update `SQUID_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## testhttpd

1. Update `BRANCH` and `TAG` files.

## unbound

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [download page](https://www.nlnetlabs.nl/projects/unbound/download/).
2. Update `UNBOUND_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## vault

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/vault/releases) and these notes:
    - https://www.vaultproject.io/docs/upgrading
    - https://www.vaultproject.io/docs/release-notes
2. Update `VAULT_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## victoriametrics

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/VictoriaMetrics/releases).
2. Check upstream `Makefile` and `Dockerfile`, and update our `Dockerfile` if needed.
   - https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/Makefile
   - https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/app/\*/Makefile
   - https://github.com/VictoriaMetrics/VictoriaMetrics/blob/vX.Y.Z/app/\*/deployment/Dockerfile
3. Update `VICTORIAMETRICS_SINGLE_VERSION` and `VICTORIAMETRICS_CLUSTER_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## victoriametrics-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/operator/releases).
2. Check upstream Makefile and Dockerfile, and update our Dockerfile if needed.
3. Update `VICTORIAMETRICS_OPERATOR_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
