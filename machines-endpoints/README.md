machines-endpoints container
============================

`machines-endpoints` is a one-shot program to create/update Kubernetes EndpointSlice and ConfigMap objects based on the information in [sabakan](https://github.com/cybozu-go/sabakan) on bootservers.

The EndpointSlice objects managed by this program are provided for [Prometheus](https://prometheus.io/) to discover services on host machines.
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

3. Check `all-servers-targets-0` EndpointSlice, `boot-servers-targets-0` EndpointSlice, `bmc-reverse-proxy` ConfigMap, and `bmc-log-collector` ConfigMap.

   ```console
   kubectl get endpointslice -n NAMESPACE all-servers-targets-0
   kubectl get endpointslice -n NAMESPACE boot-servers-targets-0
   kubectl get configmap -n NAMESPACE bmc-reverse-proxy
   kubectl get configmap -n NAMESPACE bmc-log-collector
   ```
 
Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/machines-endpoints)
