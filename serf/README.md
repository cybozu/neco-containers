serf-container
===============

[serf](https://www.serf.io) is a decentralized solution for service discovery and orchestration that is lightweight, highly available, and fault tolerant.

This repository provides a Dockerfile to build a container image for Serf.

Usage
-----

Prepare the following [Serf Configuration file](https://www.serf.io/docs/agent/options.html#configuration-files)

```json
{
  "node_name": "localhost",
  "rpc_addr": "0.0.0.0:7373",
  "tags": {
    "role": "load-balancer",
    "datacenter": "east"
  }
}
```

To launch serf server by `docker run`:

    $ docker run -d --rm --read-only --name serf \
       --mount type=bind,source=/your/config,target=/serf/config \
       -p 7373:7373 -p 7946:7946 \
       quay.io/cybozu/serf:latest \
         agent -config-dir=/serf/config

To use serf cli, first install it in a host OS directory `DIR`:

    $ docker run --rm -u root:root \
      --entrypoint /usr/local/serf/install-tools \
      --mount type=bind,source=DIR,target=/host \
      quay.io/cybozu/serf:latest

Then run `serf` as follows:

    $ DIR/serf members
 
Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/serf)
