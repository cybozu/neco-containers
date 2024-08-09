local-pv-provisioner
====================

`local-pv-provisioner` is a custom controller that creates [local](https://kubernetes.io/docs/concepts/storage/volumes/#local) PersistentVolume(PV) resources from devices that match the specified conditions. It also cleanup the PVs when it's released and with the periodical trigger.

* The PVs will be removed along with the deletion of the node because of using `ownerReferences`.

## How `local-pv-provisioner` works

`local-pv-provisioner` operates based on a configmap specified in the annotations of the node.
We'll refer to this ConfigMap as the "PV spec Configmap" throughout this text.

First, `local-pv-provisioner` checks for an annotation named `local-pv-provisioner.cybozu.io/pv-spec-configmap` 
on the node. This annotation should have the name of the PV spec ConfigMap.
If this annotation is not specified, `local-pv-provisioner` will use the value provided by the command-line argument
`--default-pv-spec-configmap`. If neither the annotation nor the command-line argument is specified,
`local-pv-provisioner` will stop working.

Next, `local-pv-provisioner` fetches the content of the specified PV spec ConfigMap and looks for
the following values:

- `deviceDir`: The directory where `local-pv-provisioner` searches for devices.
- `deviceNameFilter`: The regular expression used to filter the devices.
- `volumeMode`: The mode of the PV created by `local-pv-provisioner`, which should be either "Filesystem" or "Block".
- `fsType`: The type of the filesystem, which should be set if and only if `volumeMode` is `"Filesystem"`. Currently, `local-pv-provisioner` only supports ext4 for this field.

After obtaining these values, `local-pv-provisioner` searches for devices, based on `deviceDir` and `deviceNameFilter`.
It then creates one PV for each found device, using the
`volumeMode` and `fsType` values specified in the PV spec ConfigMap.

### An example

Let's consider the following PV spec ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pv-spec-cm-fs
data:
  volumeMode: Filesystem
  fsType: ext4
  deviceDir: /dev/disk/by-path/
  deviceNameFilter: ".*"
```

If we use this as a PV spec ConfigMap, all the devices under `/dev/disk/by-path/` should be selected.
And each created PV should be formatted as a ext4 filesystem.

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

1. Start a minikube cluster:
    ```
    $ make -C e2etest launch-cluster MINIKUBE_PROFILE=test
    ```

2. Create some loop devices for PVs:
    ```
    $ make -C e2etest create-loop-dev
    ```

3. Start `local-pv-provisioner`:
    ```
    $ make -C e2etest launch-local-pv-provisioner
    ```

4. Annotate the node:
    ```
    # If you'd like to deploy Block PVs, use pv-spec-cm-block instead.
    $ kubectl annotate node minikube-worker local-pv-provisioner.cybozu.io/pv-spec-configmap=pv-spec-cm-fs
    ```

6. Check that the pods have started and PVs have been created.
    ```
    $ kubectl get pod,pv -n kube-system
    NAME                                          READY   STATUS    RESTARTS      AGE
    pod/coredns-5d78c9869d-t6n55                  1/1     Running   0             13m
    pod/etcd-minikube-worker                      1/1     Running   0             13m
    pod/kube-apiserver-minikube-worker            1/1     Running   0             13m
    pod/kube-controller-manager-minikube-worker   1/1     Running   0             13m
    pod/kube-proxy-xkq4d                          1/1     Running   0             13m
    pod/kube-scheduler-minikube-worker            1/1     Running   0             13m
    pod/local-pv-provisioner-4ltjn                1/1     Running   0             8m3s
    pod/storage-provisioner                       1/1     Running   2 (11m ago)   13m

    NAME                                           CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM   STORAGECLASS    REASON   AGE
    persistentvolume/local-minikube-worker-loop0   1Gi        RWO            Retain           Available           local-storage            62s
    persistentvolume/local-minikube-worker-loop1   1Gi        RWO            Retain           Available           local-storage            61s
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
      volumeMode: Filesystem
      resources:
        requests:
          storage: 1Ki
    ```
    Set the values according to PV as follows:
    * `spec.storageClassName`: `local-storage`
    * `spec.accessModes`: `ReadWriteOnce`
    * `spec.volumeMode`: `Filesystem`

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
          volumeMounts:
            - name: sample-volume
              mountPath: /mnt/test-vol
      volumes:
        - name: sample-volume
          persistentVolumeClaim:
            claimName: sample-pvc
    ```

3. The PVC will be bound to a PV.
    ```
    $ kubectl get pv,pvc
    NAME                                           CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                STORAGECLASS    REASON   AGE
    persistentvolume/local-minikube-worker-loop0   1Gi        RWO            Retain           Bound       default/sample-pvc   local-storage            4m22s
    persistentvolume/local-minikube-worker-loop1   1Gi        RWO            Retain           Available                        local-storage            4m21s

    NAME                               STATUS   VOLUME                        CAPACITY   ACCESS MODES   STORAGECLASS    AGE
    persistentvolumeclaim/sample-pvc   Bound    local-minikube-worker-loop0   1Gi        RWO            local-storage   38s
    ```

### Stop the test environment

Please use `make -C e2etest clean`.

## How to cleanup released PVs

The cleanup process is:
1. Watches Update events for Persistent Volume
2. If `spec.storageClassName: local-storage` and `status.phase: Released`, fill the first 100MB of the corresponding device with zero value.
3. Delete the Persistent Volume from Kubernetes API server.
  - Note that, this process is executed even if failed to cleanup the device.

Note that this cleanup process is also executed periodically (interval: 1 hour).

## Command-line flags and environment variables

| Flag             | Env name            | Default | Description                                                                        |
| ---------------- | ------------------- | ------- | ---------------------------------------------------------------------------------- |
| metrics-addr     | LP_METRICS_ADDR     | `:8080` | Bind address for the metrics endpoint.                                             |
| node-name        | LP_NODE_NAME        |         | The name of Node on which this program is running. It is a required flag.          |
| polling-interval | LP_POLLING_INTERVAL | `5m`    | Polling interval to check devices.                                                 |
| development      | LP_DEVELOPMENT      | `false` | Use development logger config.                                                     |
| namespace-name   | LP_NAMESPACE_NAME   |         | The name of the namespace in which this program is running. It is a required flag. |

## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/local-pv-provisioner)
