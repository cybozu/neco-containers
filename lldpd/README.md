[![CircleCI](https://circleci.com/gh/cybozu/lldpd-container.svg?style=svg)](https://circleci.com/gh/cybozu/lldpd-container)
[![Docker Repository on Quay](https://quay.io/repository/cybozu/lldpd/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/lldpd)

lldpd container
===============

This repository provides a Dockerfile to build a Docker container
that runs [lldpd](https://vincentbernat.github.io/lldpd/).

Features
--------

* lldpd 1.0 based on Ubuntu 18.04.
* Multi-stage build to minimize the container size.
* Reduced functions for simplicity; SNMP and XML are disabled.

Usage
-----

### Run the container

For docker:
```
$ docker run -d --read-only --hostname=$(uname -n) \
    --cap-drop ALL --cap-add SYS_CHROOT --cap-add SETGID --cap-add NET_ADMIN --cap-add NET_RAW \
    --network host --name lldpd \
    --mount type=tmpfs,destination=/run/lldpd \
    quay.io/cybozu/lldpd:1.0
```

For rkt:
```
$ sudo rkt run \
    --hostname=$(uname -n) \
    --volume run,kind=empty,readOnly=false \
    --net=host \
    quay.io/cybozu/lldpd:1.0 \
      --readonly-rootfs=true \
      --caps-retain=CAP_SYS_CHROOT,CAP_NET_ADMIN,CAP_SETGID,CAP_NET_RAW \
      --name lldpd \
      --mount volume=run,target=/run/lldpd
```

### Use client tools

`lldpcli` is an interactive client:

```
$ docker exec -it lldpd lldpcli
[lldpcli] # show neighbors
-------------------------------------------------------------------------------
LLDP neighbors:
-------------------------------------------------------------------------------
Interface:    eth0, via: LLDP, RID: 1, Time: 0 day, 00:01:02
  Chassis:     
    ChassisID:    mac 42:01:0a:80:00:03
    SysName:      host-vm
    SysDescr:     Debian GNU/Linux 9 (stretch) Linux 4.9.0-6-amd64 #1 SMP Debian 4.9.82-1+deb9u3 (2018-03-02) x86_64
    MgmtIP:       10.128.0.3
    MgmtIP:       fe80::4001:aff:fe80:3
    Capability:   Bridge, on
    Capability:   Router, on
    Capability:   Wlan, off
    Capability:   Station, off
  Port:        
    PortID:       mac a6:f9:14:06:2f:67
    PortDescr:    vm-1_int-net
    TTL:          120
-------------------------------------------------------------------------------
```

License
-------

MIT
