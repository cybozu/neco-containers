[![Docker Repository on Quay](https://quay.io/repository/cybozu/local-pv-provisioner/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/local-pv-provisioner)

local-pv-provisioner
====================

`local-pv-provisioner` is a custom controller that creates [local](https://kubernetes.io/docs/concepts/storage/volumes/#local) PersistentVolume(PV) resources from devices that match the specified conditions.

* The PVs are linked to a node by `ownerReferences` setting.
* The PVs will be removed along with the deletion of the node.

## How to discover devices

`local-pv-provisioner` searches for devices according to `--device-path` and `--device-name-filter` options.

* `--device-path` option specifies the path to search the devices.
* `--device-name-filter` option filter the device names using a regular expression.

If you specifies the following condition, all devices under `/dev/disk/by-path/` will be selected.

```console
$ local-pv-provisioner --device-path="/dev/disk/by-path/" --device-name-filter=".*"
```

## What resources are created

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-node1-pci-0000-3b-00.0-ata-1
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
    path: /dev/disk/by-path/pci-0000:3b:00.0-ata-1
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - node1
```

* `metadata.name` is decided according to the following rules.
   * The name is a concatenation of `local`, node name, and device name with `-`.
   * If it contains characters other than alphabets, numbers, `-` and `.`, it is replaced with `-`.
* `spec.storageClassName` is automatically set a value `local-storage`.
* `spec.capacity.storage` is decided from the max capacity of the device.

## Prometheus metrics

`local-pv-provisioner` exposes the following metrics.

### `local_pv_provisioner_available_devices`

`local_pv_provisioner_available_devices` is a gauge that indicates the number of available devices recognized by `local-pv-provisioner`.

| Label  | Description            |
| ------ | ---------------------- |
| `node` | The node resource name |

### `local_pv_provisioner_error_devices`

`local_pv_provisioner_available_devices` is a gauge that indicates the number of error devices recognized by `local-pv-provisioner`.

| Label  | Description            |
| ------ | ---------------------- |
| `node` | The node resource name |

## Command-line flags and environment variables

| Flag               | Env name              | Default              | Description                                                                                         |
| ------------------ | --------------------- | -------------------- | --------------------------------------------------------------------------------------------------- |
| metrics-addr       | LP_METRICS_ADDR       | `:8080`              | Bind address for the metrics endpoint.                                                              |
| device-dir         | LP_DEVICE_DIR         | `/dev/disk/by-path/` | Path to the directory that stores the devices for which PersistentVolumes are created.              |
| device-name-filter | LP_DEVICE_NAME_FILTER | `.*`                 | A regular expression that allows selection of devices on device-idr to be created PersistentVolume. |
| node-name          | LP_NODE_NAME          |                      | The name of Node on which this program is running. It is a required flag.                           |
| polling-interval   | LP_POLLING_INTERVAL   | `10s`                | Polling interval to check devices.                                                                  |
| development        | LP_DEVELOPMENT        | `false`              | Use development logger config.                                                                      |

## Installation

```console
$ kubectl apply -f ./install.yaml
```
