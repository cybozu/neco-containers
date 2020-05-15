[![Docker Repository on Quay](https://quay.io/repository/cybozu/vault/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/vault)

vault-container
===============

[Vault](https://www.vaultproject.io) is a tool for managing secrets provided by HashiCorp.

This repository provides a Dockerfile to build a container image for Vault.

Usage
-----

Prepare the following [Vault Configuration file](https://www.vaultproject.io/docs/configuration/index.html)

```
storage "inmem" {}
listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = 1
}
```

To launch vault server by `docker run`:

    $ docker run -d --rm --read-only --name vault \
       --ulimit memlock=-1 \
       --mount type=bind,source=/your/config.hcl,target=/vault/config/config.hcl \
       -p 8200:8200 -p 8201:8201 \
       quay.io/cybozu/vault:0.10 \
         server -config=/vault/config/config.hcl

To use vault cli, first install it in a host OS directory `DIR`:

    $ docker run --rm -u root:root \
      --entrypoint /usr/local/vault/install-tools \
      --mount type=bind,source=DIR,target=/host \
      quay.io/cybozu/vault:0.10

Then run `vault` as follows:

    $ DIR/vault status
 
Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/vault)
