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

- [admission (neco-admission)](#admission-neco-admission)
- [alertmanager](#alertmanager)
- [alloy](#alloy)
- [argo-rollouts](#argo-rollouts)
- [argocd](#argocd)
- [argocd-image-updater](#argocd-image-updater)
- [bird](#bird)
- [blackbox\_exporter](#blackbox_exporter)
- [bmc-log-collector](#bmc-log-collector)
- [bmc-reverse-proxy](#bmc-reverse-proxy)
- [bpf-map-pressure-exporter](#bpf-map-pressure-exporter)
- [bpftrace](#bpftrace)
- [cadvisor](#cadvisor)
- [cep-checker](#cep-checker)
- [ceph](#ceph)
  - [Create a patched image from the specific version](#create-a-patched-image-from-the-specific-version)
- [ceph-extra-exporter](#ceph-extra-exporter)
- [cephcsi](#cephcsi)
- [cert-manager](#cert-manager)
- [chrony](#chrony)
- [cilium](#cilium)
- [cilium-certgen](#cilium-certgen)
- [cilium-operator-generic](#cilium-operator-generic)
- [configmap-reload](#configmap-reload)
- [contour](#contour)
- [coredns](#coredns)
- [csi sidecars/csi-addons-controller](#csi-sidecarscsi-addons-controller)
- [dex](#dex)
- [envoy](#envoy)
- [etcd](#etcd)
- [external-dns](#external-dns)
- [fluent-bit](#fluent-bit)
- [golang-all (golang for combinations of versions and platforms)](#golang-all-golang-for-combinations-of-versions-and-platforms)
- [gorush](#gorush)
- [grafana](#grafana)
- [grafana-operator](#grafana-operator)
- [haproxy](#haproxy)
- [heartbeat](#heartbeat)
- [hubble](#hubble)
- [hubble-relay](#hubble-relay)
- [hubble-ui](#hubble-ui)
- [kube-metrics-adapter](#kube-metrics-adapter)
- [kube-state-metrics](#kube-state-metrics)
- [kube-storage-version-migrator](#kube-storage-version-migrator)
- [kubernetes](#kubernetes)
- [local-pv-provisioner](#local-pv-provisioner)
- [loki](#loki)
- [machines-endpoints](#machines-endpoints)
- [memcached](#memcached)
- [memcached\_exporter](#memcached_exporter)
- [meows-dctest-runner](#meows-dctest-runner)
- [meows-neco-runner](#meows-neco-runner)
- [opentelemetry-collector](#opentelemetry-collector)
- [pause](#pause)
- [pomerium](#pomerium)
- [prometheus-adapter](#prometheus-adapter)
- [prometheus-config-reloader](#prometheus-config-reloader)
- [promtail](#promtail)
- [promtail-debug](#promtail-debug)
- [pushgateway](#pushgateway)
- [pyroscope](#pyroscope)
- [redis](#redis)
- [registry](#registry)
- [rook](#rook)
- [s3gw](#s3gw)
- [sealed-secrets](#sealed-secrets)
- [serf](#serf)
- [squid](#squid)
- [squid-exporter](#squid-exporter)
- [stakater/Reloader](#stakaterreloader)
- [tcp-keepalive](#tcp-keepalive)
- [teleport-node](#teleport-node)
- [tempo](#tempo)
- [testhttpd](#testhttpd)
- [trust-manager](#trust-manager)
- [trust-packages](#trust-packages)
- [unbound](#unbound)
- [unbound\_exporter](#unbound_exporter)
- [vault](#vault)
- [victoriametrics](#victoriametrics)
- [victoriametrics-datasource](#victoriametrics-datasource)
- [victoriametrics-operator](#victoriametrics-operator)

---

## admission (neco-admission)

![Kubernetes Update](./kubernetes_update.svg)

In Kubernetes update:

1. Update the following version variables in `Makefile`.
   - `CONTROLLER_TOOLS_VERSION`
   - `KUSTOMIZE_VERSION`
   - `ENVTEST_K8S_VERSION`
2. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
3. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/admission
   make generate manifests
   # Commit, if there are any updated files.
   ```

4. Confirm build and test are green.

   ```bash
   make build test
   ```

5. Update `TAG` file.

![Regular Update](./regular_update.svg)

In Regular update, do the following as part of the update of each CRD-providing product:

1. Update a matching version variable from the following in `Makefile`.
   - `CONTOUR_VERSION`
   - `ARGOCD_VERSION`
   - `GRAFANA_OPERATOR_VERSION`
2. Modify the code to match the new CRDs if CRDs are changed.
   - The code which depended on the CRDs are in the [hook](https://github.com/cybozu/neco-containers/tree/main/admission/hooks) directory.
   - And let's use `Unstructured` instead of use golang library. Take a look at [this PR](https://github.com/cybozu/neco-containers/pull/339/files).
3. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/admission
   make clean
   make generate manifests
   # Commit, if there are any updated files.
   ```

4. Confirm build and test are green.

   ```bash
   make build test
   ```

5. Update `TAG` file.

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
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Dockerfile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/alloy/blob/vX.Y.Z/tools/build-image/Dockerfile`
3. Update `ALLOY_VERSION` in `Dockerfile`.
4. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
5. Update `TAG` file.

## argo-rollouts

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/argoproj/argo-rollouts/releases) for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - `https://github.com/argoproj/argo-rollouts/blob/vX.Y.Z/Dockerfile`
3. Update version variables in `Dockerfile`.
    - Update `ARGO_ROLLOUTS_VERSION`.
4. Update `TAG` file.

## argocd

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/argoproj/argo-cd/releases) for changes.
2. Check `hack/tool-versions.sh` for the tools versions.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/hack/tool-versions.sh`
3. Update tool versions in `Dockerfile`.
    - [Kustomize](https://github.com/kubernetes-sigs/kustomize/releases)
    - [Helm](https://github.com/helm/helm/releases)
4. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/Dockerfile`
5. Update version variables in `Dockerfile`.
    - Update `ARGOCD_VERSION`, `KUSTOMIZE_VERSION` and `HELM_VERSION`.
6. Update `BRANCH` and `TAG` files.
7. Follow maintenance instructions for [neco-admission](./maintenance.md#admission-neco-admission) if needed.

> [!Note]
> ArgoCD depends on dex,Redis,HAProxy.
> So browse the following manifests and update [dex](#dex),[redis](#redis),[haproxy](#haproxy) images next.

- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`
- `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

> [!Note]
> ArgoCD's Application objects are validated by [neco-admission](#admission-neco-admission).
> If Application CRD has been changed, you may need to update [neco-admission](#admission-neco-admission).

## argocd-image-updater

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/argoproj-labs/argocd-image-updater/releases) for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
    - `https://github.com/argoproj-labs/argocd-image-updater/blob/vX.Y.Z/Dockerfile`
3. Update version variables in `Dockerfile`.
    - Update `ARGOCD_IMAGE_UPDATER_VERSION`.
4. Update `TAG` file.

## bird

![Regular Update](./regular_update.svg)

1. Check the latest version on the [download page](https://bird.network.cz/?download) of the official website.
2. Check the `https://gitlab.nic.cz/labs/bird/-/blob/vX.Y.Z/NEWS` in the official GitLab repository at the specified version tag for release notes.
3. Check the release announcement mail from the [mailing list archive](https://bird.network.cz/?m_list) on the official website.
4. Update `BIRD_VERSION` variable in `Dockerfile`.
5. Update image tag in `README.md`.
6. Update `BRANCH` and `TAG` files.

## blackbox_exporter

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/prometheus/blackbox_exporter/releases).
2. Update `BLACKBOX_EXPORTER_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` files.

## bmc-log-collector

![Regular Update](./regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Confirm test are pass.

   ```bash
   make setup
   make test
   ```

3. Update `BRANCH` and `TAG` files.

## bmc-reverse-proxy

![Kubernetes Update](./kubernetes_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Confirm test are green.

   ```bash
   make test
   ```

3. Update image tag in `bmc-reverse-proxy.yaml`.
4. Update `TAG` file.

## bpf-map-pressure-exporter

![Regular Update](./regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Upgrade base images in `Dockerfile`.
3. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

## bpftrace

![Regular Update](./regular_update.svg)

1. Check the bpftrace [release page](https://github.com/bpftrace/bpftrace/releases).
2. Check the bpftrace [Dockerfile](https://github.com/bpftrace/bpftrace/blob/master/docker/Dockerfile.ubuntu) of the latest release version.
3. Check the bcc [release page](https://github.com/iovisor/bcc/releases) and update `BCC_TARGET` in the `Makefile` to its latest version.
4. Check the libbpf [release page](https://github.com/libbpf/libbpf/releases) and update `LIBBPF_TARGET` in the `Makefile` to its latest version.
5. Update `TAG` file.

## cadvisor

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/google/cadvisor/releases).
2. Check the upstream build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/Makefile`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/build/release.sh`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/build/build.sh`
   - `https://github.com/google/cadvisor/blob/vX.Y.Z/deploy/Dockerfile`
3. Update `CADVISOR_VERSION` in `Dockerfile`
4. Update `TAG` file.

## cep-checker

![Regular Update](./regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update cilium and cilium-cli version in `e2e/Makefile` and `go.mod` to the version used by neco.
3. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

## ceph

![CSA Update](./csa_update.svg)

1. Check the [release page](https://docs.ceph.com/en/latest/releases/).
2. Check the [build ceph](https://docs.ceph.com/en/latest/install/build-ceph/) document and [README.md](https://github.com/ceph/ceph/blob/main/README.md).
   1. If other instructions are needed for `ceph/build.sh`, add the instructions.
   2. If there are ceph runtime packages or required tool changes, update Dockerfile.
3. Update the `version` argument on the `build-ceph` job in the `build_ceph` job in the Github Actions `main` workflow.
4. Update `BRANCH` and `TAG` files.

### Create a patched image from the specific version

When you want to create a new image with patches to the specific version of Ceph,
follow these steps.

1. Create a branch with the name `ceph-vX.Y.Z` from the commit you want, and push it.
   - You must follow the branch naming convention to activate the image build and push jobs.
   - If the branch already exists, you can skip this step.
2. Create a PR to the branch `ceph-vX.Y.Z`, and merge it.

## ceph-extra-exporter

![CSA Update](./csa_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Upgrade base images in `Dockerfile`.
3. Update the `TAG` files accordingly.

## cephcsi

![CSA Update](./csa_update.svg)

1. See [Rook's values.yaml file](https://github.com/rook/rook/blob/master/deploy/charts/rook-ceph/values.yaml) of the appropriate tag and check the version of cephcsi.
2. Update `CSI_IMAGE_VERSION` in Dockerfile with the value which you checked in the previous step.
3. Update `BASE_IMAGE` in Dockerfile if necessary.
   - If `BASE_IMAGE` is too old, the build may fail.
   - You should also check `BASE_IMAGE` in [the upstream build.env](https://github.com/ceph/ceph-csi/blob/devel/build.env) file of the appropriate tag.
   - If you are going to update Rook later, you should also check `CEPH_VERSION` in [the Rook's upstream Makefile](https://github.com/rook/rook/blob/master/images/ceph/Makefile) file of the appropriate tag.
   - If the desired version of ceph image does not exist in our container repository, build the ceph image first.
4. See [the upstream Dockerfile](https://github.com/ceph/ceph-csi/blob/devel/deploy/cephcsi/image/Dockerfile) of the appropriate tag, and update our Dockerfile if necessary.
5. Update `BRANCH` and `TAG` files.

> [!Note]
> Because cephcsi container is build based on the ceph container, build the ceph container first if necessary.

## cert-manager

![Regular Update](./regular_update.svg)

1. Check [releases](https://github.com/jetstack/cert-manager/releases) for changes.
2. Check whether manually applied patches have been included in the new release and remove them accordingly.
   1. If patches are still needed, synchronize the forked repository (<https://github.com/cybozu-go/cert-manager>).
   2. Create and checkout a new branch named `vX.Y.Z-neco` from the tag named `vX.Y.Z`.
   3. Cherry-pick the commit included patches and create a new tag named `vX.Y.Z-neco-longtimeout.1`.
   4. Push it.
3. Update `CERT_MANAGER_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## chrony

![Regular Update](./regular_update.svg)

1. Check the [release note](https://chrony.tuxfamily.org/news.html).
2. Update `CHRONY_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## cilium

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Update `CILIUM_IMAGE_TOOLS_TARGET` in `neco-containers/cilium/Makefile`.
   1. `CILIUM_IMAGE_TOOLS_TARGET` is a commit SHA of <https://github.com/cilium/image-tools> .
      Update to the latest one and go backward to find an appropriate SHA.
3. Checkout `cilium/cilium` and `cilium/image-tools` at the relevant SHA.
   1. Run `make checkout` and download them under `neco-containers/cilium/src`.
4. Check the upstream `Dockerfile`s to make necessary changes for `neco-containers/cilium`.
   1. Run `make urls`. It displays all the URLs of the upstream `Dockerfile`s.
   2. All the build specification is written in `neco-containers/cilium/Dockerfile`. Please check the header comment of the file to find the mapping of our build targets and the upstream ones.
5. Build `ghcr.io/cybozu/cilium` and see the result.
   1. Run `make build` to build.
   2. Run `make test` to make sanity check.
   3. Run `make test-e2e` to run the standard connectivity test.
      1. Remove `--test=!check-log-errors` in `e2e/Makefile` to confirm no important error is ignored.
      2. Restore the flag until we catch up with the upstream that solves the LLVM-17-related issue.
   4. If any problem found, `dive ghcr.io/cybozu/cilium:$(cat TAG)` will help.

> [!Note]
> The cilium-operator-generic and hubble-relay images should be updated at the same time as the cilium image for consistency.

## cilium-certgen

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/certgen/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/certgen/blob/vX.Y.Z/Dockerfile`
3. Update the `BRANCH` and `TAG` files accordingly.

## cilium-operator-generic

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/operator/Dockerfile`
3. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The cilium-operator-generic image should be updated at the same time as the cilium image for consistency.

## configmap-reload

![Regular Update](./regular_update.svg)

1. Check the [tags page](https://github.com/jimmidyson/configmap-reload/tags).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/jimmidyson/configmap-reload/blob/vX.Y.Z/Dockerfile`
3. Update `CONFIGMAP_RELOAD_VERSION` in `Dockerfile`
4. Update `BRANCH` and `TAG` files.

## contour

![Regular Update](./regular_update.svg)

> [!Note]
> Contour uses Envoy as a "data plane." Keep version correspondence between the contour and [envoy](#envoy) images. Check the compatibility matrix below.
>
> - [Contour Compatibility Matrix](https://projectcontour.io/resources/compatibility-matrix/)

1. Check the [release page](https://github.com/projectcontour/contour/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/projectcontour/contour/blob/vX.Y.Z/Dockerfile`
3. Update `CONTOUR_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.
6. Follow maintenance instructions for [neco-admission](./maintenance.md#admission-neco-admission) if needed.

> [!Note]
> Contour's HTTPProxy objects are validated by [neco-admission](#admission-neco-admission).
> If HTTPProxy CRD has been changed, you may need to update [neco-admission](#admission-neco-admission).

## coredns

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/coredns/coredns/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/coredns/coredns/blob/vX.Y.Z/Dockerfile`
3. Update `COREDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

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
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/dex/argocd-dex-server-deployment.yaml`

1. Check the [release page](https://github.com/dexidp/dex/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/dexidp/dex/blob/vX.Y.Z/Dockerfile`
3. Update `DEX_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## envoy

![Regular Update](./regular_update.svg)

> [!Note]
> Envoy is managed by [Contour](#contour) so update to the supported version. See the below.
>
> - [Contour Compatibility Matrix](https://projectcontour.io/resources/compatibility-matrix/)

1. Check the [release page](https://github.com/envoyproxy/envoy/releases).
2. Update `clang_archive_path` in [`.github/workflows/main.yaml`](/.github/workflows/main.yaml) if you want to update the clang version.
3. Update image tag in `README.md`.
4. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
5. Update `BRANCH` and `TAG` files.

## etcd

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/etcd-io/etcd/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/etcd-io/etcd/blob/vX.Y.Z/Dockerfile-release.amd64`
3. Update `ETCD_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## external-dns

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/external-dns/releases).
2. Check the upstream `.ko.yaml`. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes-sigs/external-dns/blob/vX.Y.Z/.ko.yaml`
3. Update `EXTERNALDNS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `TAG` file.

## fluent-bit

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/fluent/fluent-bit/releases).
2. Update `FLUENT_BIT_VERSION` in `Dockerfile`.
3. Update `TAG`.

## golang-all (golang for combinations of versions and platforms)

![Regular Update](./regular_update.svg)

Automated by `.github/workflows/update.yaml`.

Manual update

1. Check the [release history](https://golang.org/doc/devel/release.html).
2. Update `GO_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG`.

## gorush

Ignore!!!

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
6. Follow maintenance instructions for [neco-admission](./maintenance.md#admission-neco-admission) if needed.

> [!Note]
> Grafana Operator's GrafanaDashboard objects are validated by [neco-admission](#admission-neco-admission).
> If GrafanaDashboard CRD has been changed, you may need to update [neco-admission](#admission-neco-admission).

## haproxy

![Regular Update](./regular_update.svg)

> [!Note]
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/ha/install.yaml`

1. Check the release notes in the [official site](https://www.haproxy.org/).
   - v2.6.x: <https://github.com/docker-library/haproxy/blob/master/2.6/Dockerfile>
2. Update `HAPROXY_SHA256` in `Dockerfile`, SHA256 hash in <http://www.haproxy.org/download>
3. Update `BRANCH` and `TAG` files.

## heartbeat

![Regular Update](./regular_update.svg)

Only the base image and module dependency should be updated.

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

## hubble

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/hubble/releases) page for changes.
2. Update the `BRANCH` and `TAG` files accordingly.

Hubble image is no longer built by the upstream. If failing to build the image, check upstream changes.

## hubble-relay

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/cilium/releases) page for changes.
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cilium/cilium/blob/vX.Y.Z/images/hubble-relay/Dockerfile`
3. Update the `BRANCH` and `TAG` files accordingly.

> [!Note]
> The hubble-relay image should be updated at the same time as the cilium image for consistency.

## hubble-ui

![Regular Update](./regular_update.svg)

1. Check the [releases](https://github.com/cilium/hubble-ui/releases) page for changes.
2. Update the `BRANCH` and `TAG` files accordingly.
3. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - <https://github.com/cilium/hubble-ui/blob/master/Dockerfile>
   - Update `NODE_VERSION` and `NGINX_VERSION` in `Dockerfile`.
   - Update `NGINX_COMMIT_HASH` in `Makefile`.
      - Browse <https://github.com/nginx/docker-nginx-unprivileged/commits/main/> .
      - `NGINX_COMMIT_HASH` should be the one referencing the commit "Update mainline NGINX to <NGINX_VERSION>".
   - Run `make clean checkout`.
4. Check the upstream [Dockerfile](https://github.com/nginx/docker-nginx-unprivileged/blob/main/Dockerfile-debian.template) for unprivileged version of nginx.

   ```sh
   OLD_NGINX_VERSION=
   NEW_NGINX_VERSION=

   diff --side-by-side -W300 \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${OLD_NGINX_VERSION}/Dockerfile-debian.template) \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${NEW_NGINX_VERSION}/Dockerfile-debian.template)

   diff --side-by-side -W300 \
   <(cat hubble-ui/Dockerfile) \
   <(curl -sL https://raw.githubusercontent.com/nginx/docker-nginx-unprivileged/refs/tags/${NEW_NGINX_VERSION}/Dockerfile-debian.template)
   ```

5. Update `NJS_VERSION` and `PKG_RELEASE` in `Dockerfile`.

## kube-metrics-adapter

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/zalando-incubator/kube-metrics-adapter/releases).
2. Update `KMA_VERSION` in `Dockerfile`.
3. Update `TAG` file.

## kube-state-metrics

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes/kube-state-metrics/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes/kube-state-metrics/blob/vX.Y.Z/Dockerfile`
3. Update `KUBE_STATE_METRICS_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `TAG` files.

## kube-storage-version-migrator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/kubernetes-sigs/kube-storage-version-migrator/releases).
2. Check the upstream build files. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/Makefile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/initializer/Dockerfile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/migrator/Dockerfile`
   - `https://github.com/kubernetes-sigs/kube-storage-version-migrator/blob/vX.Y.Z/cmd/trigger/Dockerfile`
3. Update `MIGRATOR_VERSION` in `Dockerfile`
4. Update `TAG` file.

## kubernetes

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/kubernetes/kubernetes/releases).
2. Update `K8S_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

## local-pv-provisioner

![CSA Update](./csa_update.svg)

1. Update version variables in `Makefile` and `e2etest/Makefile`.
2. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
3. Generate code and manifests.

   ```bash
   cd $GOPATH/src/github.com/cybozu/neco-containers/local-pv-provisioner
   make generate manifests
   # Commit, if there are any updated files.
   ```

4. Confirm build and test are green.

   ```bash
   make build test
   ```

5. Update `TAG` file.

## loki

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/loki/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/grafana/loki/blob/vX.Y.Z/cmd/loki/Dockerfile`
3. Update `LOKI_VERSION` in `Dockerfile`.
4. Update `TAG` file.

> [!Note]
> Keep the version of [promtail](#promtail) the same as that of loki.

## machines-endpoints

![Kubernetes Update](./kubernetes_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Confirm test is green.

   ```bash
   make test
   ```

3. Update image tag in `machines-endpoints.yaml`.
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

## meows-dctest-runner

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/cybozu-go/meows/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/Dockerfile`
3. Update `MEOWS_VERSION` in `Dockerfile`.
4. Update `GO_VERSION` and `PLACEMAT_VERSION` in `Dockerfile`, if there are any updates.
   - `GO_VERSION`: <https://github.com/cybozu/neco-containers/blob/main/golang-all>
   - `PLACEMAT_VERSION`: <https://github.com/cybozu-go/placemat/releases/latest>
5. Update `BRANCH` and `TAG` files.

## meows-neco-runner

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/cybozu-go/meows/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/cybozu-go/meows/blob/vX.Y.Z/Dockerfile`
3. Update the `Dockerfile` to install the same tools as ubuntu-debug.
   - Also update `GRPCURL_VERSION`, if there are any changes.
   - <https://github.com/cybozu/ubuntu-base/blob/main/22.04/ubuntu-debug/Dockerfile#L6>
4. Update `MEOWS_VERSION` in `Dockerfile`.
5. Update `BRANCH` and `TAG` files.

## opentelemetry-collector

![Regular Update](./regular_update.svg)

opentelemetry-collector container consists of three repositories: opentelemetry-collector, opentelemetry-collector-contrib and opentelemetry-collector-releases

1. Check the release pages [main](https://github.com/open-telemetry/opentelemetry-collector/releases) [contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases) [release](https://github.com/open-telemetry/opentelemetry-collector-releases/releases).
2. Check the upstream Dockerfile and builder manifest. If there are any updates, update our `Dockerfile`.
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/Dockerfile`
   - `https://github.com/open-telemetry/opentelemetry-collector-releases/blob/vX.Y.Z/distributions/otelcol/manifest.yaml`
3. Update `OTELCOL_VERSION` and/or `OTELCOL_CONTRIB_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## pause

![Kubernetes Update](./kubernetes_update.svg)

1. Check the changelog.
   - `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/CHANGELOG.md`
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/kubernetes/kubernetes/blob/vX.Y.Z/build/pause/Dockerfile`
3. Update `K8S_VERSION` and `PAUSE_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## pomerium

![Regular Update](./regular_update.svg)

1. Check the release page and upgrade guide.
   - <https://github.com/pomerium/pomerium/releases>
   - <https://www.pomerium.com/docs/core/upgrading>
2. Check the diff of the Dockerfile.

   ```bash
   cd /path/to/pomerium
   git switch --detach v${NewVersion}
   git diff v${CurrentVersion} Dockerfile
   ```

3. Update `Dockerfile`.
   - Pomerium version
   - Golang version
   - Node.js version
4. Update `TAG` file.

## prometheus-adapter

![Regular Update](./regular_update.svg)

1. Check the release page.
   - <https://github.com/kubernetes-sigs/prometheus-adapter/releases>
2. Update version variables in `Dockerfile`.
3. Update `TAG` file.

## prometheus-config-reloader

![Regular Update](./regular_update.svg)

This is a part of [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator/).
This is used as a sidecar of Alloy Pod now.

1. Check the latest release of `prometheus-operator`
2. Update version variable in `Dockerfile`.
3. Update `TAG` file.

## promtail

![Regular Update](./regular_update.svg)

Promtail contains two versions, one for promtail and the other for libsystemd.
The promtail version should be the same with [loki](#loki).
The libsystemd version should be the same with the one running on [the stable Flatcar OS](https://www.flatcar.org/releases).

1. Update `LOKI_VERSION` in `Dockerfile`.
2. Update `SYSTEMD_VERSION` in `Dockerfile` if needed.
3. Update `TAG` file.

## promtail-debug

TBD

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

## redis

![Regular Update](./regular_update.svg)

> [!Note]
> This image is used by [ArgoCD](#argocd). So browse the following manifest and check the required version.
> If the manifest uses version _a.b.c_, we should use version _a.b.d_ where _d >= c_. Don't use a newer minor version.
>
> - `https://github.com/argoproj/argo-cd/blob/vX.Y.Z/manifests/base/redis/argocd-redis-deployment.yaml`

1. Check the release notes in the [official site](https://redis.io/).
2. Check the Dockerfile in docker-library. If there are any updates, update our `Dockerfile`.
   - v7.0.x: <https://github.com/docker-library/redis/blob/master/7.0/Dockerfile>
3. Update `REDIS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## registry

![Regular Update](./regular_update.svg)

1. Check the release notes in the [release page](https://github.com/distribution/distribution/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - <https://github.com/docker/distribution/blob/master/Dockerfile>
3. Update `REGISTRY_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## rook

![CSA Update](./csa_update.svg)

> [!Note]
> A specific version of rook depends on specific versions of csi sidecar containers listed below. Update these containers at the same time.

- cephcsi
- csi-attacher
- csi-node-driver-registrar
- csi-provisioner
- csi-resizer
- csi-snapshotter

1. Check the [release page](https://github.com/rook/rook/releases).
2. Check the `CEPH_VERSION` of the upstream Makefile. If the version is changed, update the base image version in our `Dockerfile`.
   - `https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Makefile`
   - If the desired version of ceph image does not exist in our container repository, build the ceph image first.
3. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/rook/rook/blob/vX.Y.Z/images/ceph/Dockerfile`
4. Update build image tag in Dockerfile if necessary.
5. Update `ROOK_VERSION` in `Dockerfile`.
6. Update `BRANCH` and `TAG` files.

> [!Note]
> You may choose the latest docker image for the build, regardless of the upstream go version.
> The current go compiler builds with the language version and toolchain version based on the go version specified in the go.mod file.
> There is no need to use an older version of the image to match go.mod.

## s3gw

![Regular Update](./regular_update.svg)

Only the base image and module dependency should be updated.

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

## sealed-secrets

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/bitnami-labs/sealed-secrets/releases).
2. Check the upstream Dockerfile and compare with ours especially on the runtime stage. If there are any updates, update our `Dockerfile`.
    - `https://github.com/bitnami-labs/sealed-secrets/blob/vX.Y.Z/docker/controller.Dockerfile`
    - `https://github.com/bitnami-labs/sealed-secrets/blob/vX.Y.Z/docker/kubeseal.Dockerfile`
3. Update `SEALED_SECRETS_VERSION` in `Dockerfile`.
4. Update `BRANCH` and `TAG` files.

## serf

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/serf/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/hashicorp/serf/blob/vX.Y.Z/scripts/serf-builder/Dockerfile`
3. Update `SERF_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## squid

![Regular Update](./regular_update.svg)

1. Check the latest **stable** version at <http://www.squid-cache.org/Versions/>
2. Check release notes if a new version is released.
    - e.g., `https://github.com/squid-cache/squid/releases/tag/SQUID_X_Y`
3. Update `SQUID_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## squid-exporter

![Regular Update](./regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update squid version in `Makefile` and `e2e/pod.yaml` if there are any updates.
3. Update `TAG` by incrementing the patch revision, e.g. 1.0.1, 1.0.2, ...

> [!Note]
> The squid images should be updated at the same time as the squid-exporter image for consistency.

## stakater/Reloader

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/stakater/Reloader/releases).
2. Check the upstream Dockerfile. If there are any updates, update our `Dockerfile`.
   - `https://github.com/stakater/Reloader/blob/vX.Y.Z/Dockerfile`
3. Update `BRANCH` and `TAG` files.

## tcp-keepalive

TBD

## teleport-node

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/gravitational/teleport/releases).
2. Run `make -C teleport-node/ check-teleport-update` and check the upstream `Makefile` and `version.mk`.
3. Update tools version in `Dockerfile`.
4. Update `Dockerfile` If there are any changes to the build method.
5. Update `TELEPORT_VERSION` in `Dockerfile`.
6. Update `TAG` files.

## tempo

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/grafana/tempo/releases).
2. Check the upstream `Makefile` and `cmd/tempo//Dockerfile`. If they have been updated significantly, update our `Dockerfile`.
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/Makefile`
   - `https://github.com/grafana/tempo/blob/vX.Y.Z/cmd/tempo/Dockerfile`
3. Update `TEMPO_VERSION` in `Dockerfile`.
4. Update `TAG` file.

## testhttpd

![Regular Update](./regular_update.svg)

1. Upgrade direct dependencies listed in `go.mod`. Use `go get` or your editor's function.
2. Update `BRANCH` and `TAG` files.

## trust-manager

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/cert-manager/trust-manager/releases).
2. Update `BRANCH` and `TAG` files.

## trust-packages

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/cert-manager/trust-manager/releases).
2. Update `TRUST_MANAGER_VERSION` in `Dockerfile`.
3. Update `TAG` file.
4. Verify that the new `trust-packages` version matches the `trust-manager` version.

## unbound

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [download page](https://www.nlnetlabs.nl/projects/unbound/download/).
2. Run `make update-root-hints`.
3. Update `UNBOUND_VERSION` in `Dockerfile`.
4. Update image tag in `README.md`.
5. Update `BRANCH` and `TAG` files.

## unbound_exporter

![Kubernetes Update](./kubernetes_update.svg)

1. Check the [release page](https://github.com/letsencrypt/unbound_exporter/releases)
2. Update `UNBOUND_EXPORTER_VERSION` in `Dockerfile`.
3. Update `BRANCH` and `TAG` files.

## vault

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/hashicorp/vault/releases) and these notes:
    - <https://www.vaultproject.io/docs/upgrading>
    - <https://www.vaultproject.io/docs/release-notes>
2. Update `VAULT_VERSION` in `Dockerfile`.
3. Update image tag in `README.md`.
4. Update `BRANCH` and `TAG` files.

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

## victoriametrics-datasource

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/victoriametrics-datasource/releases).
2. Check upstream Makefile and Dockerfile, and update our Dockerfile if needed.
3. Update `VERSION` in `Dockerfile`.
4. Update `TAG` file.

## victoriametrics-operator

![Regular Update](./regular_update.svg)

1. Check the [release page](https://github.com/VictoriaMetrics/operator/releases).
2. Check upstream Makefile and Dockerfile, and update our Dockerfile if needed.
3. Update `VICTORIAMETRICS_OPERATOR_VERSION` in `Dockerfile`.
4. Update `TAG` file.
