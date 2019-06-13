[![Docker Repository on Quay](https://quay.io/repository/cybozu/csi-sidecar/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/csi-sidecar)

CSI sidecar container
=====================

This directory provides a Dockerfile to build a Docker container that contains as follows.
- [external-provisioner](https://github.com/kubernetes-csi/external-provisioner) 
- [node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar)
- [external-attacher](https://github.com/kubernetes-csi/external-attacher)

Usage
-----

### Start `csi-sidecar`

Run the container

```console
# Run as external-provisioner
$ docker run -d --read-only --name=csi-provisioner \
    --entrypoint=csi-provisioner \
    quay.io/cybozu/csi-sidecar:1.1.1.1 \
    --csi-address=/run/topolvm/csi-topolvm.sock \
    --feature-gates=Topology=true

# Run as node-driver-registrar
$ docker run -d --read-only --name=csi-node-driver-registrar \
    --entrypoint=csi-node-driver-registrar \
    quay.io/cybozu/csi-sidecar:1.1.1.1 \
    --csi-address=/run/topolvm/csi-topolvm.sock \
    --kubelet-registration-path=/var/lib/kubelet/plugins/topolvm.cybozu.com/node/csi-topolvm.sock 

# Run as external-attacher
$ docker run -d --read-only --name=csi-attacher \
    --entrypoint=csi-attacher \
    quay.io/cybozu/csi-sidecar:1.1.1.1 \
    --csi-address=/run/topolvm/csi-topolvm.sock
```
