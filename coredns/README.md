[![Docker Repository on Quay](https://quay.io/repository/cybozu/coredns/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/coredns)

# CoreDNS container

[CoreDNS](https://coredns.io/) is DNS server typically used on Kubernetes.

## Usage

To launch server with specific config file.

    $ docker run quay.io/cybozu/coredns:1.3 -v Corefile:/etc/coredns/Corefile -- \
        -conf /etc/coredns/Corefile
