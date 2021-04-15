# CoreDNS container

[CoreDNS](https://coredns.io/) is DNS server typically used on Kubernetes.

## Usage

To launch server with specific config file.

    $ docker run quay.io/cybozu/coredns:1.8 -v Corefile:/etc/coredns/Corefile -- \
        -conf /etc/coredns/Corefile
 
## Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/coredns)
