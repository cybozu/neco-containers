[BIRD][] container
==================

This directory provides a Dockerfile to build a Docker container
that runs up-to-date [BIRD][] internet routing daemon.

Features
--------

* BIRD 2.x
* Multi-stage build to minimize the container size.
* Optimized for BGP.  RIP, OSPF, Babel, and RAdv are not built-in.

Usage
-----

### Prepare `bird.conf`

See http://bird.network.cz/?get_doc&v=20&f=bird-3.html

### Run the container

For docker:
```
$ docker run -d --read-only --cap-drop ALL \
    --cap-add=NET_ADMIN --cap-add NET_BIND_SERVICE --cap-add NET_RAW \
    --network host --name bird \
    --mount type=tmpfs,destination=/run/bird \
    --mount type=bind,source=/your/bird.conf,target=/etc/bird/bird.conf \
    quay.io/cybozu/bird:2.14
```

### Use client tools

`birdc` is an interactive client:

```
$ docker exec -it bird birdc
BIRD 2.0.2 ready.

bird> show status
Router ID is 172.17.0.2
Current server time is 2018-04-10 05:26:11.287
Last reboot on 2018-04-10 05:25:59.011
Last reconfiguration on 2018-04-10 05:25:59.011
Daemon is up and running

bird> show memory
BIRD memory usage
Routing tables:     25 kB
Route attributes: 6224  B
Protocols:        4880  B
Total:              67 kB

bird> quit
```

`birdcl` is a light-weight client:

```
$ docker exec bird birdcl show status
BIRD 2.0.2 ready.
BIRD 2.0.2
Router ID is 10.146.0.4
Current server time is 2018-04-12 13:30:27.181
Last reboot on 2018-04-12 13:23:57.909
Last reconfiguration on 2018-04-12 13:23:57.909
Daemon is up and running
```

[BIRD]: https://bird.network.cz/

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/bird)
