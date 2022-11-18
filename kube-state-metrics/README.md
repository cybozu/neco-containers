kube-state-metrics
==================

[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is a service that listens to the Kubernetes API server and generates prometheus metrics about the state of the objects.

Usage
-----

```console
$ docker run -p 8080:8080 -p 8081:8081 \
    quay.io/cybozu/kube-state-metrics:2.6.0.1 \
    --kubeconfig=<KUBE-CONFIG>\
```

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/kube-state-metrics)
