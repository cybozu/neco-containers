[![Docker Repository on Quay](https://quay.io/repository/cybozu/kubernetes/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/kubernetes)

kubernetes container
===================

[kubernetes](https://github.com/kubernetes/kubernetes) image contains binaries for the Kubernetes components.

Contained binaries:

- kube-apiserver
- kube-controller-manager
- kube-proxy
- kube-scheduler
- kubelet

Usage
-----

To launch `apiserver` by `docker run`:

    $ docker run --net=host --name apiserver -d \
        quay.io/cybozu/kubernetes:1.17 kube-apiserver \
        --advertise-address=192.168.1.101 \
        --insecure-bind-address=0.0.0.0 \
        --insecure-port=8080 \
        --enable-bootstrap-token-auth=true \
        --etcd-servers=http://192.168.1.101:2379,http://192.168.1.102:2379,http://192.168.1.103:2379 \
        --storage-backend=etcd3
