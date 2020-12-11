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
![Regular Update](./regular_update.svg)

1. Update version variables in `Makefile`.
2. Update go modules.
   - In the regular update:
      ```bash
      $ cd $GOPATH/src/github.com/cybozu/neco-containers/admission
      $ go get github.com/projectcalico/libcalico-go@v<LIBCALICO_VERSION>
      $ go get github.com/projectcontour/contour@v<CONTOUR_VERSION>
      $ go mod tidy
      ```
   - In the Kubernetes update:
      ```bash
      $ cd $GOPATH/src/github.com/cybozu/neco-containers/admission
      $ K8SLIB_VERSION=X.Y.Z # e.g. K8SLIB_VERSION=0.18.9
      $ go get k8s.io/api@v$K8SLIB_VERSION k8s.io/apimachinery@v$K8SLIB_VERSION k8s.io/client-go@v$K8SLIB_VERSION
      $ go get sigs.k8s.io/controller-runtime@v<CTRL_VERSION>
      $ go mod tidy
      ```
3. Modify the code to match the new CRDs if CRDs are changed.
   - The code which depended on the CRDs are in the [hook](https://github.com/cybozu/neco-containers/tree/master/admission/hooks) directory.
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

## argocd

![Kubernetes Update](./kubernetes_update.svg)
![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.
2. Check the upstream's Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/argoproj/argo-cd/blob/vX.Y.Z/Dockerfile
4. Update version variables in `Dockefile`.
   - In the regular update: Update `ARGOCD_VERSION`, `KUSTOMIZE_VERSION` and `PACKR_VERSION`.
   - In the Kubernetes update: Update `KUSTOMIZE_VERSION`.
5. Update `BRANCH` and `TAG` files.

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
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/projectcalico/node/blob/vX.Y.Z/Dockerfile.amd64
   - https://github.com/projectcalico/typha/blob/vX.Y.Z/docker-image/Dockerfile.amd64
3. Update version variables (`CALICO_VERSION` and `TINI_VERSION`) in `Dockefile`.
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

## cke-tools

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/containernetworking/plugins/releases).
2. Update `CNI_PLUGIN_VERSION` in `Dockerfile`
3. Update `src/CHANGELOG.md`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## configmap-reload

1. Check the [release page](https://github.com/jimmidyson/configmap-reload/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/jimmidyson/configmap-reload/blob/vX.Y.Z/Dockerfile
3. Update `CONFIGMAP_RELOAD_VERSION` in `Dockerfile`
4. Update `src/CHANGELOG.md`.
5. Update `BRANCH` and `TAG` files.

## contour

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/projectcontour/contour/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/projectcontour/contour/blob/vX.Y.Z/Dockerfile
3. Update `CONTOUR_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## coredns

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/coredns/coredns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/coredns/coredns/blob/vX.Y.Z/Dockerfile
3. Update `COREDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## dex

![Regular Update](./regular_update.svg)

***NOTE:*** This image is used by ArgoCD. So browse the following manifest and check the required version.
- https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml

1. Check the [release page](https://github.com/dexidp/dex/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/dexidp/dex/blob/vX.Y.Z/Dockerfile
3. Update `DEX_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## dnsmasq

![Regular Update](./regular_update.svg)

1. Check the [dnsmasq-base package](https://packages.ubuntu.com/bionic/dnsmasq-base) in the Ubuntu packages.
2. Check the [CHANGELOG](http://www.thekelleys.org.uk/dnsmasq/CHANGELOG) in the official website, if the `dnsmasq-base` package is updated.
4. Update `DNSMASQ_VERSION` in `Dockerfile`.
5. Update image tag in `README.md`.
6. Update `BRANCH` and `TAG` files.

## envoy

![Regular Update](./regular_update.svg)

***NOTE:*** Envoy is managed by Contour so update to the supported version. See the below.
- [Envoy Support Matrix](https://projectcontour.io/resources/envoy/)

1. Check the [release page](https://github.com/envoyproxy/envoy/releases).
2. Update the `version` argument on the `build-envoy` job in the CircleCI `main` workflow.
3. Update `BAZEL_VERSION` in `build-envoy` job. The required version is written in the following file.
   - https://github.com/envoyproxy/envoy/blob/vX.Y.Z/.bazelversion
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## etcd

***NOTE:*** The etcd is upgraded in neither the Regular Update nor the Kubernetes Update.
Upgrading etcd needs a special procedure, so we do it as a special task.

1. Check the [release page](https://github.com/etcd-io/etcd/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/etcd-io/etcd/blob/vX.Y.Z/Dockerfile-release
3. Update `ETCD_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## external-dns

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/external-dns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/kubernetes-sigs/external-dns/blob/vX.Y.Z/Dockerfile
3. Update `EXTERNALDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `TAG` file.

## filebeat

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/elastic/beats/releases).
2. Update `BEATS_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG`.

## golang / golang-bionic

![Regular Update](./regular_update.svg)

1. Check the [release history](https://golang.org/doc/devel/release.html).
2. Update `GO_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG`.

## gorush

1. Check the [release page](https://github.com/elastic/beats/releases).
2. Update `GORUSH_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## grafana

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/grafana/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/grafana/grafana/blob/vX.Y.Z/Dockerfile
3. Update `GRAFANA_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## grafana-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/integr8ly/grafana-operator/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/integr8ly/grafana-operator/blob/vX.Y.Z/build/Dockerfile
3. Update `VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG`.

## kube-state-metrics

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes/kube-state-metrics/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
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

## mackerel-agent

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/mackerelio/mackerel-agent/releases).
2. Update `MACKEREL_AGENT_VERSION` in `Dockerfile`.
3. Update `TAG` file.

## metallb

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/metallb/metallb/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/metallb/metallb/blob/vX.Y.Z/controller/Dockerfile
   - https://github.com/metallb/metallb/blob/vX.Y.Z/speaker/Dockerfile
3. Update `METALLB_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## metrics-server

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/metrics-server/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/kubernetes-sigs/metrics-server/blob/vX.Y.Z/Dockerfile
   - NOTE: Dockerfile path is different between v0.3 and v0.4. So when upgrading from v0.3 to v0.4, check the following file.
   - https://github.com/kubernetes-sigs/metrics-server/blob/release-0.3/deploy/docker/Dockerfile
3. Update `METRICS_SERVER_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## moco-mysql

T.B.D.

## pause

![Kubernetes Update](./kubernetes_update.svg)

1. Check the changelog.
   - https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/CHANGELOG.md
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/Dockerfile
3. Update `PAUSE_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## prometheus

![Regular Update](./regular_update.svg)

This container image contains 3 components(prometheus, alertmanager and pushgateway). So please check each component.

1. Check the release page.
   - https://github.com/prometheus/prometheus/releases
   - https://github.com/prometheus/alertmanager/releases
   - https://github.com/prometheus/pushgateway/releases
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/prometheus/prometheus/blob/vX.Y.Z/Dockerfile
   - https://github.com/prometheus/alertmanager/blob/vX.Y.Z/Dockerfile
   - https://github.com/prometheus/pushgateway/blob/vX.Y.Z/Dockerfile
3. Update version variables in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## prometheus-config-reloader

T.B.D.

## redis

![Regular Update](./regular_update.svg)

***NOTE:*** This image is used by ArgoCD. So browse the following manifest and check the required version.
- https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml

1. Check the release notes in the [official site](https://redis.io/).
2. Check the Dockerfile in docker-library. If there are any updates, update our `Dockefile`.
   - v5.0: https://github.com/docker-library/redis/blob/master/5.0/Dockerfile
   - v6.0: https://github.com/docker-library/redis/blob/master/6.0/Dockerfile
3. Update `REDIS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## rook

***NOTE:*** The rook image is based on the ceph image. So upgrade the ceph image first.

1. Check the [release page](https://github.com/rook/rook/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Dockerfile
3. Check the `TINI_VERSION` in the following Makefile.
   - https://github.com/rook/rook/blob/vX.Y.Z/images/Makefile
4. Update `ROOK_VERSION` and `TINI_VERSION` in `Dockerfile`.
5. Update ceph image tag in `Dockerfile`.
6. Update `BRANCH` and `TAG` files.

## serf

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/serf/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockefile`.
   - https://github.com/hashicorp/serf/blob/vX.Y.Z/scripts/serf-builder/Dockerfile
3. Update `SERF_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## squid

![Regular Update](./regular_update.svg)

1. Check the [squid package](https://packages.ubuntu.com/bionic/squid) in the Ubuntu packages.
2. Check the [ChangeLog](https://github.com/squid-cache/squid/blob/master/ChangeLog) in the official website, if the `squid` package is updated.
3. Update `SQUID_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## tail

***NOTE:*** This image will be removed soon.

1. Check the [release page](https://github.com/nxadm/tail/releases).
2. Update go modules.
   ```bash
   $ cd $GOPATH/src/github.com/cybozu/neco-containers/tail/src
   $ go get github.com/nxadm/tail@v<TAIL_VERSION>
   $ go mod tidy
   ```
3. Confirm test are green.
   ```bash
   $ make test
   ```
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

1. Check the [release page](https://github.com/hashicorp/vault/releases) and the [ChangeLog](https://github.com/hashicorp/vault/blob/master/CHANGELOG.md).
2. Update `VAULT_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## victoriametrics

T.B.D.

## victoriametrics-operator

T.B.D.
