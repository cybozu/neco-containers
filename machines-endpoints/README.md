machines-endpoints-container
============================

`machines-endpoints` is one shot program to create kubernetes endpoints from sabakan on bootservers.

This program is for prometheus to discover services on host machines.

This program works in kubernetes pods.

Usage
-----
1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/master/docs/dctest.md) environment.
2. Deploy rbac and cronjob resources for `machines-endpoints`.
    ```console
    $ kubectl apply -n monitoring -f machines-endpoints.yaml
    ```
3. Check `prometheus-node-targets`.
   ```console
   $ kubectl -n monitoring get endpoints prometheus-node-targets
   ```
