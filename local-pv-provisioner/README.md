local-pv-provisioner
====================

`local-pv-provisioner` is a custom controller that creates [local](https://kubernetes.io/docs/concepts/storage/volumes/#local) PersistentVolume(PV) resources from devices that match the specified conditions. It also cleanup the PVs when it's released and with the periodical trigger.

* The PVs will be removed along with the deletion of the node because of using `ownerReferences`.

## How to discover devices

`local-pv-provisioner` searches for devices according to `--device-path` and `--device-name-filter` options.

* `--device-path` option specifies the path to search the devices.
* `--device-name-filter` option filters the device names using a regular expression.

If you specify the following condition, all devices under `/dev/disk/by-path/` will be selected.

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
* The device specified `spec.local.path` is cleaned up by filling the first 100MB with zero value.

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

## Usage

### Start `local-pv-provisioner`

1. Create symbolic links to device files that you want to expose for pods in `/dev/crypt-disk/by-path`.

2. Prepare kind environment.
    ```
    $ kind create cluster --config cluster.yaml --wait=300s
    ```

3. Deploy `local-pv-provisioner`.
    ```
    $ kubectl apply -f local-pv-provisioner.yaml
    ```

4. Check that the pods have started and PVs have been created.
    ```
    $ kubectl get pod,pv
    NAME                             READY   STATUS    RESTARTS   AGE
    pod/local-pv-provisioner-5kn9n   1/1     Running   0          49s
    pod/local-pv-provisioner-rq8sm   1/1     Running   0          46s

    NAME                                               CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS    REASON   AGE
    persistentvolume/local-kind-worker-dummy-dev-01    1Ki        RWO            Retain           Available           local-storage            37s
    persistentvolume/local-kind-worker-dummy-dev-02    1Ki        RWO            Retain           Available           local-storage            37s
    persistentvolume/local-kind-worker2-dummy-dev-01   1Ki        RWO            Retain           Available           local-storage            34s
    persistentvolume/local-kind-worker2-dummy-dev-02   1Ki        RWO            Retain           Available           local-storage            34s
    ```

### How to use volumes

1. Create a PVC as follows:
    ```yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: sample-pvc
    spec:
      storageClassName: local-storage
      accessModes:
        - ReadWriteOnce
      volumeMode: Block
      resources:
        requests:
          storage: 1Ki
    ```
    Set the values according to PV as follows:
    * `spec.storageClassName`: `local-storage`
    * `spec.accessModes`: `ReadWriteOnce`
    * `spec.volumeMode`: `Block`

2. Create a pod as follows:
    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: sample-pod
    spec:
      containers:
        - name: ubuntu
          image: ghcr.io/cybozu/ubuntu:20.04
          command: ["/usr/local/bin/pause"]
          volumeDevices:
            - name: sample-volume
              devicePath: /dev/sample-dev
      volumes:
        - name: sample-volume
          persistentVolumeClaim:
            claimName: sample-pvc
    ```

3. The PVC will be bound to a PV.
    ```
    $ kubectl get pv,pvc
    NAME                                               CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                STORAGECLASS    REASON   AGE
    persistentvolume/local-kind-worker-dummy-dev-01    1Ki        RWO            Retain           Available                        local-storage            3m29s
    persistentvolume/local-kind-worker-dummy-dev-02    1Ki        RWO            Retain           Available                        local-storage            3m29s
    persistentvolume/local-kind-worker2-dummy-dev-01   1Ki        RWO            Retain           Available                        local-storage            3m26s
    persistentvolume/local-kind-worker2-dummy-dev-02   1Ki        RWO            Retain           Bound       default/sample-pvc   local-storage            3m26s

    NAME                               STATUS   VOLUME                            CAPACITY   ACCESS MODES   STORAGECLASS    AGE
    persistentvolumeclaim/sample-pvc   Bound    local-kind-worker2-dummy-dev-02   1Ki        RWO            local-storage   29s
    ```

## How to cleanup released PVs

The cleanup process is:
1. Watches Update events for Persistent Volume
2. If `spec.storageClassName: local-storage` and `status.phase: Released`, fill the first 100MB of the corresponding device with zero value.
3. Delete the Persistent Volume from Kubernetes API server.
  - Note that, this process is executed even if failed to cleanup the device.

Note that this cleanup process is also executed periodically (interval: 1 hour).

## Command-line flags and environment variables

| Flag               | Env name              | Default              | Description                                                                                         |
| ------------------ | --------------------- | -------------------- | --------------------------------------------------------------------------------------------------- |
| metrics-addr       | LP_METRICS_ADDR       | `:8080`              | Bind address for the metrics endpoint.                                                              |
| device-dir         | LP_DEVICE_DIR         | `/dev/disk/by-path/` | Path to the directory that stores the devices for which PersistentVolumes are created.              |
| device-name-filter | LP_DEVICE_NAME_FILTER | `.*`                 | A regular expression that allows selection of devices on device-idr to be created PersistentVolume. |
| node-name          | LP_NODE_NAME          |                      | The name of Node on which this program is running. It is a required flag.                           |
| polling-interval   | LP_POLLING_INTERVAL   | `5m`                 | Polling interval to check devices.                                                                  |
| development        | LP_DEVELOPMENT        | `false`              | Use development logger config.                                                                      |

## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/local-pv-provisioner)
