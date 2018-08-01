[![Docker Repository on Quay](https://quay.io/repository/cybozu/vault/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/vault)

vault-container
===============

[Vault](https://www.vaultproject.io) is a tool for managing secrets provided by HashiCorp.
This repository provides a Dockerfile which contains `vault` binary.

Usage
-----

Prepare the following [Vault Configuration file](https://www.vaultproject.io/docs/configuration/index.html)

```
storage "file" {
  path = "/vault/files"
}
listener "tcp" {
  address     = "127.0.0.1:8200"
  tls_disable = 1
}
```

To launch vault server by `docker run`:

    $ docker run -d --read-only \
       --name vault-1 \
       --mount type=bind,source=/your/config.hcl,target=/vault/config/config.hcl \
       --mount type=bind,source=/your/files,target=/vault/files \
       -p 8200:8200 -p 8201:8201 \
       --cap-add=IPC_LOCK \
       quay.io/cybozu/vault:0.10.4 \
         server -config=/vault/config/config.hcl

To use vault cli by `docker run`:

    $ docker run --rm -it \
        --env VAULT_ADDR="http://127.0.0.1:8200" \
        quay.io/cybozu/vault:0.10.4 \
          list secret/
