dnsmasq-container
==============

[dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) provides network infrastructure for small networks: DNS, DHCP, router advertisement and network boot.

Usage
-----

To launch `dnsmasq` by `docker run`:

    $ docker run --rm --cap-drop ALL --cap-add=NET_ADMIN \
         --net=host quay.io/cybozu/dnsmasq:2.84 \
         -d -q \
         --dhcp-range=192.168.1.3,192.168.1.254
 
Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/dnsmasq)
