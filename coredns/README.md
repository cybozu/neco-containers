# CoreDNS container

[CoreDNS](https://coredns.io/) is DNS server typically used on Kubernetes.

## Usage

To launch server with specific config file.

    $ docker run ghcr.io/cybozu/coredns:1.11 -v Corefile:/etc/coredns/Corefile -- \
        -conf /etc/coredns/Corefile
 
## Docker images

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/coredns)
