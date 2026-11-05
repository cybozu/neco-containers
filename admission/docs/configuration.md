# Configuration

The configuration of `neco-admission` is a collection of webhooks configurations.
This collection is indexed by webhooks names.

## HTTPProxyMutator

The configuration of `HTTPProxyMutator` is a map with the following keys.

| Name     | Type       | Description       |
| -------- | ---------- | ----------------- |
| policies | \[\]policy | A list of policy. |

Each policy represents an IP restriction that you want to add.
The order of ipAllowPolicy for HTTPProxy resources modified with this mutating webhook is not guaranteed.
if the originally listed ipAllowPolicy and the ipAllowPolicy specified in the annotation are the same, the duplicate IPFilterPolicy will not be added.

| Name          | Type                                                                                                      | Description |
| ------------- | --------------------------------------------------------------------------------------------------------- | ----------- |
| name          | string                                                                                                    |             |
| ipAllowPolicy | \[\][IPFilterPolicy](https://projectcontour.io/docs/main/config/api/#projectcontour.io/v1.IPFilterPolicy) |             |

### Example

config.yaml

```yaml
HTTPProxyMutator:
  policies:
  - name: <policy name>
    ipAllowPolicy:
    - source: Peer
      cidr: xxx.yyy.zzz.111
```

httpproxy.yaml

```yaml
apiVersion: projectcontour.io/v1
kind: HTTPProxy
metadata:
  name: hoge
  namespace: fuga
  annotations:
    admission.cybozu.com/ip-policy: <policy name>
spec:
  ~~~
  routes:
  - ~~~
    # insert from mutating webhook
    ipAllowPolicy:
    - source: Peer
      cidr: xxx.yyy.zzz.111
```
