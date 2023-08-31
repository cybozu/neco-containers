machines-endpoints container
============================

`machines-endpoints` is a one-shot program to create/update Kubernetes Endpoints, EndpointSlice and ConfigMap objects based on the information in [sabakan](https://github.com/cybozu-go/sabakan) on bootservers.

The Endpoints/EndpointSlice objects managed by this program are provided for [Prometheus](https://prometheus.io/) to discover services on host machines.
* The host machines listed by this program include spare machines and boot servers.
    Such machines are not registered in Kubernetes as Nodes, and they cannot be scraped with `node` role in `<kubernetes_sd_config>` configuration.
* Retired machines are not listed because they never provide metrics.

The ConfigMap object is provided for [BMC reverse proxy](https://github.com/cybozu/neco-containers/tree/main/bmc-reverse-proxy) to resolve BMC hostnames to IP addresses.
* The host machines listed by this program include spare machines and boot servers.
* Retired machines are also listed because we need to operate them via BMCs.

This program works in kubernetes pods.

Usage
-----

1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/main/docs/dctest.md) environment.
2. Deploy RBAC and CronJob resources for `machines-endpoints`.

   ```console
   vi machines-endpoints.yaml  # adjust tag of container image to the latest one
   kubectl apply -n NAMESPACE -f machines-endpoints.yaml
   ```

3. Check `prometheus-node-targets` endpoints, `bootserver-etcd-metrics` endpoints, and `bmc-reverse-proxy` configmap.

   ```console
   kubectl get endpoints -n NAMESPACE prometheus-node-targets
   kubectl get endpointslice -n NAMESPACE prometheus-node-targets
   kubectl get endpoints -n NAMESPACE bootserver-etcd-metrics
   kubectl get endpointslice -n NAMESPACE bootserver-etcd-metrics
   kubectl get configmap -n NAMESPACE bmc-reverse-proxy
   ```
 
Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/machines-endpoints)
