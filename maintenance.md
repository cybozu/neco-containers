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

- [admission (neco-admission)](./admission/MAINTENANCE.md)
- [alertmanager](./alertmanager/MAINTENANCE.md)
- [alloy](./alloy/MAINTENANCE.md)
- [approver-policy](./approver-policy/MAINTENANCE.md)
- [argo-rollouts](./argo-rollouts/MAINTENANCE.md)
- [argocd](./argocd/MAINTENANCE.md)
- [argocd-image-updater](./argocd-image-updater/MAINTENANCE.md)
- [bird](./bird/MAINTENANCE.md)
- [blackbox\_exporter](./blackbox_exporter/MAINTENANCE.md)
- [bmc-log-collector](./bmc-log-collector/MAINTENANCE.md)
- [bmc-reverse-proxy](./bmc-reverse-proxy/MAINTENANCE.md)
- [bpftrace](./bpftrace/MAINTENANCE.md)
- [cadvisor](./cadvisor/MAINTENANCE.md)
- [cep-checker](./cep-checker/MAINTENANCE.md)
- [ceph](./ceph/MAINTENANCE.md)
- [ceph-csi-operator](./ceph-csi-operator/MAINTENANCE.md)
- [ceph-extra-exporter](./ceph-extra-exporter/MAINTENANCE.md)
- [cephcsi](./cephcsi/MAINTENANCE.md)
- [cert-manager](./cert-manager/MAINTENANCE.md)
- [chrony](./chrony/MAINTENANCE.md)
- [cilium](#cilium)
- [cilium-certgen](./cilium-certgen/MAINTENANCE.md)
- [cilium-operator-generic](#cilium-operator-generic)
- [contour](./contour/MAINTENANCE.md)
- [coredns](./coredns/MAINTENANCE.md)
- [csi sidecars/csi-addons-controller](#csi-sidecarscsi-addons-controller)
- [ct-monitor](./ct-monitor/MAINTENANCE.md)
- [dex](#dex)
- [envoy](./envoy/MAINTENANCE.md)
- [etcd](./etcd/MAINTENANCE.md)
- [external-dns](./external-dns/MAINTENANCE.md)
- [falco](./falco/MAINTENANCE.md)
- [falco-k8s-metacollector](./falco-k8s-metacollector/MAINTENANCE.md)
- [falco-operator](./falco-operator/MAINTENANCE.md)
- [falcosidekick](./falcosidekick/MAINTENANCE.md)
- [fluent-bit](./fluent-bit/MAINTENANCE.md)
- [gitleaks](./gitleaks/MAINTENANCE.md)
- [golang-all (golang for combinations of versions and platforms)](./golang-all/MAINTENANCE.md)
- [grafana](./grafana/MAINTENANCE.md)
- [grafana-operator](./grafana-operator/MAINTENANCE.md)
- [haproxy](#haproxy)
- [hubble](#hubble)
- [hubble-relay](#hubble-relay)
- [hubble-ui](./hubble-ui/MAINTENANCE.md)
- [kube-metrics-adapter](./kube-metrics-adapter/MAINTENANCE.md)
- [kube-state-metrics](./kube-state-metrics/MAINTENANCE.md)
- [kube-storage-version-migrator](./kube-storage-version-migrator/MAINTENANCE.md)
- [kubernetes](./kubernetes/MAINTENANCE.md)
- [local-pv-provisioner](./local-pv-provisioner/MAINTENANCE.md)
- [loki](./loki/MAINTENANCE.md)
- [machines-endpoints](./machines-endpoints/MAINTENANCE.md)
- [memcached](./memcached/MAINTENANCE.md)
- [memcached\_exporter](./memcached-exporter/MAINTENANCE.md)
- [meows-dctest-runner](./meows-dctest-runner/MAINTENANCE.md)
- [meows-neco-runner](./meows-neco-runner/MAINTENANCE.md)
- [neco-exporter](./neco-exporter/MAINTENANCE.md)
- [opentelemetry-collector](./opentelemetry-collector/MAINTENANCE.md)
- [pause](./pause/MAINTENANCE.md)
- [pomerium](./pomerium/MAINTENANCE.md)
- [prometheus-adapter](./prometheus-adapter/MAINTENANCE.md)
- [prometheus-config-reloader](./prometheus-config-reloader/MAINTENANCE.md)
- [pushgateway](./pushgateway/MAINTENANCE.md)
- [pyroscope](./pyroscope/MAINTENANCE.md)
- [redis](#redis)
- [registry](./registry/MAINTENANCE.md)
- [rook](./rook/MAINTENANCE.md)
- [s3gw](./s3gw/MAINTENANCE.md)
- [sealed-secrets](./sealed-secrets/MAINTENANCE.md)
- [serf](./serf/MAINTENANCE.md)
- [spegel](./spegel/MAINTENANCE.md)
- [spiffe-csi-driver](./spiffe-csi-driver/MAINTENANCE.md)
- [spiffe-test](./spiffe-test/MAINTENANCE.md)
- [spire-agent](./spire-agent/MAINTENANCE.md)
- [spire-server](./spire-server/MAINTENANCE.md)
- [spire-controller-manager](./spire-controller-manager/MAINTENANCE.md)
- [squid](./squid/MAINTENANCE.md)
- [squid-exporter](./squid-exporter/MAINTENANCE.md)
- [stakater/Reloader](./stakater-reloader/MAINTENANCE.md)
- [tcp-keepalive](./tcp-keepalive/MAINTENANCE.md)
- [teleport-node](./teleport-node/MAINTENANCE.md)
- [tempo](./tempo/MAINTENANCE.md)
- [testhttpd](./testhttpd/MAINTENANCE.md)
- [trust-manager](./trust-manager/MAINTENANCE.md)
- [trust-packages](./trust-packages/MAINTENANCE.md)
- [unbound](./unbound/MAINTENANCE.md)
- [unbound\_exporter](./unbound_exporter/MAINTENANCE.md)
- [vault](./vault/MAINTENANCE.md)
- [victoriametrics](./victoriametrics/MAINTENANCE.md)
- [victoriametrics-operator](./victoriametrics-operator/MAINTENANCE.md)

---

## cilium

![Regular Update](./regular_update.svg)

Cilium images are split into per-minor-version directories (e.g. `cilium-1.18/`).
Each `cilium-X.Y/` directory contains the images that must match the Cilium minor version:

- `cilium-X.Y/cilium`
- `cilium-X.Y/cilium-operator-generic`
- `cilium-X.Y/hubble-relay`
- `cilium-X.Y/hubble`

[cilium-certgen](./cilium-certgen/MAINTENANCE.md) and [hubble-ui](./hubble-ui/MAINTENANCE.md) are shared by all Cilium versions and live in top-level directories.

To update an existing minor version, follow the steps below inside `cilium-X.Y/`.

To add a new minor version:

1. Copy the latest `cilium-X.Y/` directory to `cilium-X.Y'/`.
2. Update the four images following each section. Update the `BRANCH` (`X.Y'`) and `TAG` files of all of them.
3. Update `scan.identifier` in `cilium/build-targets.yaml` and `hubble-relay/build-targets.yaml` to `ghcr.io/cybozu/<image>:X.Y'`.
4. Add `/cilium-X.Y'/` to `.github/CODEOWNERS`.

To remove an old minor version, check that no cluster uses it, then delete `cilium-X.Y/` and its CODEOWNERS entry.

> [!Note]
> Changes that are not specific to a Cilium version (e.g. pinning base images, CI-related settings) should be applied to all `cilium-X.Y/` directories in the same PR.

To update the cilium image:

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Update `CILIUM_IMAGE_TOOLS_TARGET` in `cilium-X.Y/cilium/Makefile`. This is a commit SHA of <https://github.com/cilium/image-tools> .
   1. Run `make IMAGE_TAG=quay.io/cilium/cilium:vX.Y.Z test` to see tools versions upstream uses.
   2. See [history](https://github.com/cilium/image-tools/commits/master/) of `image-tools` to find an appropriate SHA, that provides matching tools version (especially for LLVM/Clang).
3. Checkout `cilium/cilium` and `cilium/image-tools` at the relevant SHA.
   1. Run `make checkout` and download them under `cilium-X.Y/cilium/src`.

> [!Note]
> `CILIUM_TARGET` in `cilium-X.Y/cilium/Makefile` pins a specific commit SHA of the `cybozu-go/cilium` fork branch (e.g. `vX.Y.Z-cybozu`), not the branch tip.
> Changes to that branch must go through a pull request. After merging one (e.g. additional cherry-picked patches), you must update `CILIUM_TARGET` to the new commit SHA; it will not be picked up automatically.
> In `cilium-1.17`, `make checkout-cilium` (and therefore `make checkout`/`make build`) runs `check-cilium-target` first, which fails if `CILIUM_TARGET` is not the current tip of that branch, to catch this case.
4. Check the upstream `Dockerfile`s to make necessary changes for `cilium-X.Y/cilium`.
   1. Run `make urls`. It displays all the URLs of the upstream `Dockerfile`s.
   2. All the build specification is written in `cilium-X.Y/cilium/Dockerfile`. Please check the header comment of the file to find the mapping of our build targets and the upstream ones.
   3. Update github.com/go-delve/delve/cmd/dlv version. Upstream may be using @latest but we must specify a version.
5. Update the cilium-cli version used by the e2e test. See cilium-cli's [README](https://github.com/cilium/cilium-cli?tab=readme-ov-file#releases) to make sure that the cilium-cli version is compatible with the cilium version.
   - Cilium 1.17: `CILIUM_CLI_VERSION` in `e2e/Makefile`.
   - Cilium 1.18 or later: `cilium/cilium-cli` in `aqua.yaml`, then run `aqua update-checksum -prune`.
6. Build `ghcr.io/cybozu/cilium` and see the result.
   1. Run `make build` to build.
   2. Run `make test` to make sanity check.
   3. Run `make test-e2e` to run the standard connectivity test.
      - `check-log-errors` has often been disabled to work around minor known issues.
      Check the logs for significant errors manually and consider re-enabling it.
   4. If any problem found, `dive ghcr.io/cybozu/cilium:$(cat TAG)` will help.

> [!Note]
> The cilium-operator-generic and hubble-relay images should be updated at the same time as the cilium image in the same `cilium-X.Y/` directory for consistency.

## cilium-operator-generic

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/operator/Dockerfile`

> [!Note]
> `ENV CILIUM_SHA` in `neco-containers/cilium-X.Y/cilium-operator-generic/Dockerfile` pins a specific commit SHA of the `cybozu-go/cilium` fork branch (e.g. `vX.Y.Z-cybozu`), not the branch tip.
> Changes to that branch must go through a pull request. After merging one (e.g. additional cherry-picked patches), you must update `CILIUM_SHA` to the new commit SHA; it will not be picked up automatically.
> `make check-cilium-target` (run by CI before the image build) fails if `CILIUM_SHA` is not the current tip of that branch, to catch this case.

3. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The cilium-operator-generic image should be updated at the same time as the cilium image in the same `cilium-X.Y/` directory for consistency.

## csi sidecars/csi-addons-controller

![CSA Update](./csa_update.svg)

This section applies to the following containers. These containers are maintained similarly.

- csi-addons-controller
- csi-addons-sidecar
- csi-attacher
- csi-node-driver-registrar
- csi-provisioner
- csi-resizer
- csi-snapshotter

1. See [Rook's values.yaml file](https://github.com/rook/rook/blob/master/deploy/charts/rook-ceph/values.yaml) of the appropriate tag and check the version of csi sidecars. The csi-addons-controller version should be the same as the csi-addons-sidecar version.
2. Update `VERSION` in Dockerfile with the value which you checked in the previous step.
3. See the upstream Dockerfile of the appropriate tag, and update our Dockerfile if necessary. The upstream Dockerfile is listed below.
   - [csi-addons-controller](https://github.com/csi-addons/kubernetes-csi-addons/blob/main/Dockerfile)
   - [csi-addons-sidecar](https://github.com/csi-addons/kubernetes-csi-addons/blob/main/build/Containerfile.sidecar)
   - [csi-attacher](https://github.com/kubernetes-csi/external-attacher/blob/master/Dockerfile)
   - [csi-node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar/blob/master/Dockerfile)
   - [csi-provisioner](https://github.com/kubernetes-csi/external-provisioner/blob/master/Dockerfile)
   - [csi-resizer](https://github.com/kubernetes-csi/external-resizer/blob/master/Dockerfile)
   - [csi-snapshotter](https://github.com/kubernetes-csi/external-snapshotter/blob/master/cmd/csi-snapshotter/Dockerfile)
4. update image tag in `Dockerfile` if necessary.
5. Update `BRANCH` and `TAG` files.

> [!Note]
> You can choose the latest stable Ubuntu image for runtime. The upstream uses distroless as the base image for runtime, while Neco uses Ubuntu for easier debugging.
> We have not decided to use Ubuntu for csi-addons-controller and csi-addons-sidecar yet, therefore, they use distroless as the base image.

<br>

> [!Note]
> You may choose the latest docker image for the build, regardless of the upstream go version. The current go compiler builds with the language version and toolchain version based on the go version specified in the go.mod file. There is no need to use an older version of the image to match go.mod. As a known issue, the upstream build script warns that `test-gofmt and test-vendor are known to be sensitive to the version of Go.`. However, we use the latest docker image unless the test fails.

## dex

![Regular Update](./regular_update.svg)

> [!Note]
> This image is used by [ArgoCD](./argocd/MAINTENANCE.md). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`

1. Check the [release page](https://github.com/dexidp/dex/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/dexidp/dex/blob/vX.Y.Z/Dockerfile`
3. Update `DEX_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## haproxy

![Regular Update](./regular_update.svg)

> [!Note]
> This image is used by [ArgoCD](./argocd/MAINTENANCE.md). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

1. Check the release notes in the [official site](https://www.haproxy.org/).
   - v2.6.x: <https://github.com/docker-library/haproxy/blob/master/2.6/Dockerfile>
2. Update `HAPROXY_SHA256` in `Dockerfile`, SHA256 hash in <http://www.haproxy.org/download>
3. Update `BRANCH` and `TAG` files.

## hubble

![Regular Update](./regular_update.svg)

> [!Note]
> Keep the Hubble CLI on the same minor version as Cilium in each `cilium-X.Y/hubble`.

1. Check the [releases](https://github.com/cilium/hubble/releases) page for changes.
2. Update `HUBBLE_SHA` in `Dockerfile` with the commit SHA of the new release tag.
3. Update the `BRANCH` and `TAG` files accordingly.

Hubble image is no longer built by the upstream. If failing to build the image, check upstream changes.

## hubble-relay

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/hubble-relay/Dockerfile`
3. Update `CILIUM_SHA` in `Dockerfile`.
4. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The hubble-relay image should be updated at the same time as the cilium image in the same `cilium-X.Y/` directory for consistency.

## redis

![Regular Update](./regular_update.svg)

> [!Note]
> This image is used by [ArgoCD](./argocd/MAINTENANCE.md). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`

1. Check the release notes in the [official site](https://redis.io/).
2. Check the Dockerfile in docker-library. If there are any updates, update our `Dockerfile`.
   - v7.0.x: <https://github.com/docker-library/redis/blob/master/7.0/Dockerfile>
3. Update `REDIS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
5. Update `README.md`, `COPYING` and the container LABEL in terms of license if necessary.

## redis_exporter

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/oliver006/redis_exporter/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
3. Update `REDIS_EXPORTER_VERSION` and checksum of archive in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.
