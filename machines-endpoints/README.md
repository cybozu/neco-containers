machines-endpoints container
============================

`machines-endpoints` is a one-shot program to create/update Kubernetes Service, EndpointSlice, and ConfigMap objects based on the information in [sabakan](https://github.com/cybozu-go/sabakan) on bootservers.

The Service and EndpointSlice objects managed by this program are provided for [Prometheus](https://prometheus.io/) to discover services on host machines.
* The Service is a headless Service (`ClusterIP: None`) that only groups the ports exposed on the target; the EndpointSlice(s) associated with it via the `kubernetes.io/service-name` label carry the actual endpoint addresses.
* The host machines listed by this program include boot servers and old-style spare machines.
    Such machines are not registered in Kubernetes as Nodes, and they cannot be scraped with `node` role in `<kubernetes_sd_config>` configuration.
    Note that recent versions of CKE configure spare machines as Kubernetes Nodes.
* Retired machines are not listed because they never provide metrics.

The `bmc-reverse-proxy` ConfigMap object is provided for [BMC reverse proxy](https://github.com/cybozu/neco-containers/tree/main/bmc-reverse-proxy) to resolve BMC hostnames to IP addresses.
* The host machines listed by this program include spare machines and boot servers.
* Retired machines are also listed because we need to operate them via BMCs.

The `bmc-log-collector` ConfigMap object is provided for [bmc-log-collector](https://github.com/cybozu/neco-containers/tree/main/bmc-log-collector) as the "machineslist.json" it reads to know which BMCs to collect hardware logs from (serial, BMC IP, and node IP for each machine).
* The host machines listed by this program include spare machines and boot servers.
* Retired machines are also listed because we need to operate them via BMCs.

This program works in kubernetes pods.

Usage
-----

1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/main/docs/dctest.md) environment.
2. Deploy RBAC and CronJob resources for `machines-endpoints`.

   ```console
   vi machines-endpoints.yaml  # adjust tag of container image and --sabakan-address to the latest/actual ones
   kubectl apply -n NAMESPACE -f machines-endpoints.yaml
   ```

3. Check the generated resources.

   ```console
   kubectl get service -n NAMESPACE all-servers-targets
   kubectl get service -n NAMESPACE boot-servers-targets
   kubectl get endpointslice -n NAMESPACE all-servers-targets-0
   kubectl get endpointslice -n NAMESPACE boot-servers-targets-0
   kubectl get configmap -n NAMESPACE bmc-reverse-proxy
   kubectl get configmap -n NAMESPACE bmc-log-collector
   ```

Options
-------

| Option | Default | Description |
| ------ | ------- | ----------- |
| `--sabakan-address` | (required) | Address of sabakan's GraphQL API, in the form `host:port`. A hostname is accepted; prefer a name that resolves to every boot server, or a VIP, so that one unreachable boot server does not stop updates. |
| `--all-servers-port` | (none) | Port to expose on the `all-servers-targets` target, which lists all non-retired machines. In the form `port:name`, repeatable. The Service and EndpointSlices for the target are created only when at least one port is given. |
| `--boot-servers-port` | (none) | Same as `--all-servers-port`, but for the `boot-servers-targets` target, which lists non-retired boot servers only. |
| `--max-endpoints-per-slice` | `100` | Maximum number of endpoints per EndpointSlice. Must be between 1 and 1000. |
| `--bmc-reverse-proxy-configmap` | `false` | Generate the `bmc-reverse-proxy` ConfigMap. |
| `--bmc-log-collector-configmap` | `false` | Generate the `bmc-log-collector` ConfigMap. |
| `--dry-run` | `false` | Print the JSON of the resources that would be applied to stdout instead of applying them. A usable kubeconfig is still required, and EndpointSlices that would be deleted are not reported. |

Labels
------

Each Service is labeled with:

| Label | Value |
| ----- | ----- |
| `app.kubernetes.io/managed-by` | `machines-endpoints.cybozu.io` |
| `app.kubernetes.io/name` | The Service's name (e.g. `all-servers-targets`) |

Each EndpointSlice is labeled with:

| Label | Value |
| ----- | ----- |
| `endpointslice.kubernetes.io/managed-by` | `machines-endpoints.cybozu.io` |
| `kubernetes.io/service-name` | The name of the Service it belongs to |

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/machines-endpoints)
