[![Docker Repository on Quay](https://quay.io/repository/cybozu/local-pv-provisioner/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/local-pv-provisioner)

local-pv-provisioner
====================

`local-pv-provisioner` is a custom controller that creates [local](https://kubernetes.io/docs/concepts/storage/volumes/#local) PersistentVolume resources from devices that match the specified conditions.

* Created PVs will be removed along with the deletion of the node to which it belongs.

## How to discover devices

You can specify the condition of the target devices by command-line args with a regular expression.
If you specifies the following condition, devices under `/dev/disk/by-path/` will be selected.

```console
$ local-pv-provisioner --device-path="/dev/disk/by-path/" --device-name-filter=".*"
```

## What resources are created

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: node1-sda
  ownerReferences:
    - apiVersion: v1
      kind: Node
      name: node1
      uid: a958d5a1-2644-4e26-abb3-f810576b2f7f
spec:
  capacity:
    storage: 100Gi
  volumeMode: Block
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /dev/sda
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - node1
```

## Command-line flags and environment variables

| Flag               | Env name              | Default              | Description                                                                                         |
| ------------------ | --------------------- | -------------------- | --------------------------------------------------------------------------------------------------- |
| metrics-addr       | LP_METRICS_ADDR       | `:8180`              | Bind address for the metrics endpoint.                                                              |
| device-dir         | LP_DEVICE_DIR         | `/dev/disk/by-path/` | Path to the directory that stores the devices for which PersistentVolumes are created.              |
| device-name-filter | LP_DEVICE_NAME_FILTER | `.*`                 | A regular expression that allows selection of devices on device-idr to be created PersistentVolume. |
| node-name          | LP_NODE_NAME          | `"`                  | The name of Node on which this program is running.                                                  |

## How to decide the size of PV

`pv.spec.capacity.storage` is decided from the max capacity of the disk.

## Installation

```console
$ kubectl apply -f ./install.yaml
```
