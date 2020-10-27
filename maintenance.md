How to maintain neco-containers
===============================

## Description

Aside from on-demand update, most of the images in neco-containers are updated regularly to keep up with the newest Ubuntu base image and software version.
Two procedures are used to maintain them.

## Regular update

This procedure describes the steps to update container images, that are not related to kubernetes.
It is performed once in a quarter.

1. Update [cybozu/ubuntu-base](https://github.com/cybozu/ubuntu-base).
2. Update the ubuntu base image for `golang` container.
   The Go version should not be updated in this procedure, because we decide when to update Go version separately.
3. Read the release notes of the following containers to decide if we update the software version in this procedure.
   1. If the software has a major version update, check if it is backward-compatible.
   2. If the version up is not backward-compatible, write another task to update the software. 
      In this case, update only the Ubuntu base image.
4. Build a new container image using the latest Ubuntu base image and the determined software version.
   For some containers, follow the dedicated instruction for it.
    - `argocd`
    - `bird`  
      See the [release note](https://gitlab.nic.cz/labs/bird/blob/master/NEWS).
    - `calico`
    - `cert-manager`
    - `chrony`  
      See the [release note](https://chrony.tuxfamily.org/news.html).
    - `contour`
    - `dex`  
      Update version to one that is used by ArgoCD. Browse [this file](https://github.com/argoproj/argo-cd/blob/master/manifests/base/dex/argocd-dex-server-deployment.yaml) and move to the release branch to see the version.
    - `dnsmasq`
    - `envoy`  
      Update version to one that is supported by Contour. See [Envoy Support Matrix](https://projectcontour.io/resources/envoy/).
    - `external-dns`
    - `grafana`
    - `grafana-operator`
    - `metallb`
    - `metrics-server`
    - `prometheus`
    - `redis`  
      Update version to one that is used by ArgoCD. Browse [this file](https://github.com/argoproj/argo-cd/blob/master/manifests/base/redis/argocd-redis-deployment.yaml) and move to the release branch to see the version.  
      See the [release note](https://redis.io/download).
    - `serf`
    - `squid`  
      Keep track of [the latest apt package version](https://launchpad.net/ubuntu/+source/squid3).
      This container is deprecated and will be removed sometime.
    - `vault`

### Notes
- Most of the softwares provide the release note on GitHub, but some do not.
  Their release notes are pointed in the above list.
- If only the Ubuntu base image is updated while the software version is decided to remain unchanged, increment the last number of the image tag to create the newer image like `.2` and `.3`. 
- If build fails, consult the upstream `Dockerfile` to keep up with the content.

## Kubernetes update

This procedure describes steps to update container images, that explicitly depend on kubernetes version.
It is performed once in a quarter.

1. Decide the kubernetes version to use.
2. Update container images used by CKE.
    - `admission`  
      Update `go.mod`.
    - `bmc-reverse-proxy`  
      Update `go.mod`.
    - `cke-tools`  
      Use the latest CNI plugin.  
      Update `go.mod`.
    - `coredns`  
      Upgrade to the latest version.  Handle any breaking changes.
    - `kubernetes`  
      Update to the decided kubernetes version.
    - `local-pv-provisioner`  
      Update `go.mod`.
    - `machines-endpoints`  
      Update `go.mod`.
    - `pause`  
      Follow the version in kubernetes repository, though it is rarely updated.
    - `unbound`  
      Upgrade to the latest version.  Handle any breaking changes.
3. Update container images used in [neco-apps](https://github.com/cybozu-go/neco-apps) to follow the kubernetes version.
    - `kube-state-metrics`
    - `metrics-server`
    - `kubectl` in ArgoCD container

### Notes

- `etcd` is updated in another procedure, because its update needs a dedicated instruction.

## Containers not updated regularly
- `ceph`
- `filebeat`
- `gorush`
- `moco-mysql`
- `rook`
- `tail`
- `testhttpd`
