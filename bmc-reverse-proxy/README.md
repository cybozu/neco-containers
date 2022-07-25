bmc-reverse-proxy-container
============================

`bmc-reverse-proxy` is a reverse proxy server for BMC.

The following products are assumed as BMC.

- dell iDRAC.

This program provides reverse proxy at once for the two ports (443, 5900) used by iDRAC.

And it resolve the BMC address by ConfigMap and the host name of the access destination.

This program works in kubernetes pods.

Usage
-----

### Start `bmc-reverse-proxy`

1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/main/docs/dctest.md) environment.
2. Run [neco-apps test](https://github.com/cybozu-private/neco-apps/blob/main/test/README.md) to setup [cert-manager](https://github.com/jetstack/cert-manager).
3. Before running `bmc-reverse-proxy`, create ConfigMap using `machines-endpoints`.
   Please read [README.md](../machines-endpoints/README.md) at machines-endpoints directory and apply a yaml file like bellow.

   ```console
   kubectl apply -f machines-endpoints.yaml
   ```

4. Once ConfigMap has been created, stop `machines-endpoints` and modify the ConfigMap to point a service which listens on TCP 443, e.g. `teleport-proxy`.
   This is because the current BMC implementation of placemat does not provide Web interfaces.

   ```console
   kubectl get configmaps bmc-reverse-proxy
   (wait for success)
   
   kubectl delete cronjobs machines-endpoints-cronjob

   kubectl get -n teleport teleport-proxy
   (check CLUSTER-IP)
   
   kubectl edit configmaps bmc-reverse-proxy
   (add a line of "teleport: <CLUSTER-IP>" in "data")
   ```

5. Deploy bmc-reverse proxy.

   ```console
   kubectl apply -f bmc-reverse-proxy.yaml
   ```

6. Check that you can access BMC via proxy.

   ```console
   kubectl get services bmc-reverse-proxy
   (check EXTERNAL-IP of proxy)

   sudo vi /etc/hosts
   (add a line of "<EXTERNAL-IP> teleport.bmc.gcp0.dev-ne.co")
   
   curl -k https://teleport.bmc.gcp0.dev-ne.co
   ```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/bmc-reverse-proxy)
