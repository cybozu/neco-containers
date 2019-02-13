[![Docker Repository on Quay](https://quay.io/repository/cybozu/kube-state-metrics/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/kube-state-metrics)

kube-state-metrics
==================

_kube-state-metrics_ is a service that listens to the Kubernetes API server and generates prometheus metrics about the state of the objects.

Usage
-----

Run `kube-state-metrics`

```console
$ docker run -p 8080:8080 -p 8081:8081 \
    quay.io/cybozu/kube-state-metrics:1.5.0.1 \
    --kubeconfig=<KUBE-CONFIG>\
    --apiserver=<APISERVER>
```
