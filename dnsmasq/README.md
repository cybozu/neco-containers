[![Docker Repository on Quay](https://quay.io/repository/cybozu/dnsmasq/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/dnsmasq)

dnsmasq-container
==============

[dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) provides network infrastructure for small networks: DNS, DHCP, router advertisement and network boot.

Usage
-----

To launch `dnsmasq` by `docker run`:

    $ docker run --rm --cap-add=NET_ADMIN --net=host quay.io/cybozu/dnsmasq \
         -d -q \
         --dhcp-range=192.168.1.3,192.168.1.254
