machines-endpoints-container
============================

`machines-endpoints` is one shot program to create kubernetes endpoints and configmap from sabakan on bootservers.

This program is (1) for prometheus to discover services on host machines and (2) for BMC proxy to resolve BMC hostnames to IP addresses.

This program works in kubernetes pods.

Usage
-----

1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/main/docs/dctest.md) environment.
2. Deploy RBAC and CronJob resources for `machines-endpoints`.

   ```console
   vi machines-endpoints.yaml  # adjust tag of container image to the latest one
   kubectl apply -f machines-endpoints.yaml
   ```

3. Check `prometheus-node-targets` endpoints, `bootserver-etcd-metrics` endpoints, and `bmc-proxy` configmap.

   ```console
   kubectl get endpoints prometheus-node-targets
   kubectl get endpoints bootserver-etcd-metrics
   kubectl get configmap bmc-proxy
   ```
 
Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/machines-endpoints)
