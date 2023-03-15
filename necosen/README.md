# necosen

`necosen` (neco-sentinel) is an [external auth server](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter) for envoy.
Currently, it can check the source IP of incoming packets.

## Settings

To restrict the source IP of incoming packets, install necosen with the following ConfigMap.

```yaml
apiVersion: v1
metadata:
  name: necosen-config
data:
  config.yaml: |
    sourceIP:
      allowedCIDRs:
        - 10.0.0.0/24
```

## Integrate with Contour

necosen can be used with [Contour](https://projectcontour.io/).
Define `spec.virtualhost.authorization.extensionRef` as follows.  
Note that [only https virtual hosts are supported](https://projectcontour.io/guides/external-authorization/).

```yaml
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: frontend
spec:
  virtualhost:
    fqdn: frontend.example.com
    tls:
      secretName: frontend
    authorization:
      extensionRef:
        name: necosen
        namespace: contour
  routes:
  - services:
    - name: backend
      port: 80
```
