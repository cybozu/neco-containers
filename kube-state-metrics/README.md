kube-state-metrics
==================

[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is a service that listens to the Kubernetes API server and generates prometheus metrics about the state of the objects.

Usage
-----

```console
$ docker run -p 8080:8080 -p 8081:8081 \
    ghcr.io/cybozu/kube-state-metrics:2.15.0.1 \
    --kubeconfig=<KUBE-CONFIG>\
```

Docker images
-------------

Docker images are available on [ghcr.io](ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/kube-state-metrics)
