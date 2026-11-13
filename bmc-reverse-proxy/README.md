bmc-reverse-proxy-container
============================

`bmc-reverse-proxy` is a reverse proxy server for BMC.

The following products are assumed as BMC.

- dell iDRAC.

This program provides reverse proxy at once for the two ports (443, 5900) used by iDRAC.

It resolves the BMC address from the host name of the access destination, using the resolve map file.

This program works in kubernetes pods.

The container image also contains `update-resolve-map`, which writes the resolve map file
based on the information in [sabakan](https://github.com/cybozu-go/sabakan).
It runs as a (native) sidecar of `bmc-reverse-proxy`, and they share the file via an `emptyDir` volume.
See [update-resolve-map](#update-resolve-map) for details.

Usage
-----

### Start `bmc-reverse-proxy`

1. Prepare [dctest](https://github.com/cybozu-go/neco/blob/main/docs/dctest.md) environment.
2. Run [neco-apps test](https://github.com/cybozu-private/neco-apps/blob/main/test/README.md) to setup [cert-manager](https://github.com/jetstack/cert-manager).
3. Edit `-sabakan-address` and the domain of the `Certificate` (`*.bmc.example.com`) in `bmc-reverse-proxy.yaml`, and deploy bmc-reverse-proxy.
   The manifest requires Kubernetes 1.30 or later, because it runs `update-resolve-map` as a native sidecar
   and uses the `sleep` action in the `preStop` hook.

   ```console
   vi bmc-reverse-proxy.yaml
   kubectl apply -f bmc-reverse-proxy.yaml
   ```

4. Check that the resolve map file has been written.

   ```console
   kubectl logs deploy/bmc-reverse-proxy -c update-resolve-map
   (check "updated resolve map" is logged)
   ```

5. Check that requests are forwarded to the BMC address.

   ```console
   kubectl get services bmc-reverse-proxy
   (check EXTERNAL-IP of proxy)

   curl -k --resolve rack0-cs1.bmc.example.com:443:<EXTERNAL-IP> https://rack0-cs1.bmc.example.com
   ```

   The BMCs of placemat do not provide Web interfaces, so the request fails with 502.
   Check that the `proxy error` log of `bmc-reverse-proxy` shows the BMC address of the machine in `backend`.
   If the request fails with 404 and `failed to resolve hostname` is logged instead, the hostname is not in the resolve map.

   ```console
   kubectl logs deploy/bmc-reverse-proxy -c bmc-reverse-proxy
   ```

Options
-------

### `bmc-reverse-proxy`

| Option         | Default                                 | Description                                                                   |
| -------------- | --------------------------------------- | ----------------------------------------------------------------------------- |
| `-tls-cert`    | `/etc/tls/tls.crt`                      | Path of the TLS certificate file.                                             |
| `-tls-key`     | `/etc/tls/tls.key`                      | Path of the TLS private key file.                                             |
| `-resolve-map` | `/var/lib/resolve-map/resolve-map.json` | Path of the resolve map file written by `update-resolve-map`.                 |
| `-no-tls`      | `false`                                 | Serve plain HTTP without loading the TLS certificate. For local testing only. |

`bmc-reverse-proxy` re-reads the TLS certificate and the resolve map file every minute.
Until the resolve map file is loaded for the first time, it retries reading the file every second.
If the resolve map file cannot be read, it logs an error and keeps using the last loaded map.
While waiting for the file to be written for the first time, it logs `waiting for resolve map` instead of an error.
If the TLS certificate cannot be read, it exits.

Requests for a hostname that is not in the resolve map are not forwarded.
`bmc-reverse-proxy` returns 404 and logs `failed to resolve hostname` at WARN level,
so that they can be distinguished from failures of BMCs (502).

### Probes

`bmc-reverse-proxy` serves the following endpoints for the probes of Kubernetes on port 8080.

| Path | Description |
| ---- | ----------- |
| `/healthz` | Returns 200 while the process is running. For liveness probe. |
| `/readyz` | Returns 200 after the resolve map file and the TLS certificate (unless `-no-tls`) have been loaded, and 503 before that. For readiness probe. |

Until they are loaded, the Pod is not ready and receives no traffic via the Service.
Once loaded, it stays ready even if the file cannot be read later, because the last loaded map is kept.

### Local testing with `-no-tls`

By default, `bmc-reverse-proxy` loads the TLS certificate from `/etc/tls/tls.crt` and `/etc/tls/tls.key` (changeable with `-tls-cert` and `-tls-key`), and exits if they cannot be read.
When you run it on your machine to check its behavior, you can use `-no-tls` to skip loading the certificate.
With this option, the proxy serves plain HTTP on the same ports (8443 and 5900).
The connection from the proxy to the BMC still uses HTTPS.

Do not use this option in production because the connection from clients is not encrypted.

1. Write a resolve map file, which maps hostnames to addresses.
   The keys are case-insensitive.
   The address should be a server that listens on HTTPS (443 for port 8443, 5900 for port 5900).

   ```console
   echo '{"teleport": "<ADDRESS>"}' > /tmp/resolve-map.json
   ```

2. Run `bmc-reverse-proxy` with `-no-tls`.

   ```console
   go run ./cmd/bmc-reverse-proxy -no-tls -resolve-map /tmp/resolve-map.json
   ```

3. Access `http://<HOSTNAME>.localhost:8443/` with a browser or curl.
   The first label of the hostname is used to look up the resolve map,
   so you cannot use `localhost` or an IP address as is.

   ```console
   curl http://teleport.localhost:8443/
   ```

   Major browsers (Chrome, Firefox, Edge) and curl resolve names ending with `.localhost` to the loopback address by themselves,
   so you do not need to edit `/etc/hosts`.
   Note that other tools using the OS resolver may not resolve them.

   To access the proxy from another machine, add the hostname to `/etc/hosts` of that machine instead,
   e.g. `192.0.2.10 teleport.bmc.local`, and access `http://teleport.bmc.local:8443/`.

You can also write the file with `update-resolve-map` if sabakan is reachable from your machine.

```console
go run ./cmd/update-resolve-map -sabakan-address <SABAKAN_ADDRESS>:10080 -output /tmp/resolve-map.json -interval 0
```

update-resolve-map
------------------

`update-resolve-map` writes the resolve map file, which maps hostnames to BMC addresses.
It gets the machines from sabakan's GraphQL API and writes the file every `-interval`.
The file is replaced atomically, so `bmc-reverse-proxy` never reads a partially written file.
If sabakan cannot be reached, or no BMC addresses are found in the response of sabakan,
it logs an error and keeps the last written file.
Until the file is written for the first time, it retries every 5 seconds instead of every `-interval`.

The file is a JSON object like below.

```json
{
  "198-51-100-196": "192.0.2.4",
  "ABC1234": "192.0.2.4",
  "rack1-cs4": "192.0.2.4"
}
```

The keys are exactly the names that clients can use.
`bmc-reverse-proxy` takes the first label of the host of a request, e.g. `rack1-cs4` in `rack1-cs4.bmc.<cluster>.<base>`,
and looks it up in the file case-insensitively, because hostnames are case-insensitive.
It does no other conversion, so all the rules of naming are in `update-resolve-map`.
Hostnames are written in lower case as in DNS, and serials in upper case as shown by sabakan and BMCs.

For each machine that has both a node address and a BMC address, the following keys are registered.
Retired machines are also included because we need to operate them via BMCs.

| Key | Example | Description |
| --- | ------- | ----------- |
| Hostname | `rack1-cs4`, `boot-0` | `rack<RACK>-<ROLE><INDEX>` in lower case. For boot servers, `boot-<RACK>` without the cluster name (e.g. not `stage0-boot-0`). |
| Node address | `198-51-100-196` | The first node IPv4 address with dots replaced by hyphens, to match the wildcard `*.bmc.<cluster>.<base>`. |
| Serial | `ABC1234` | The serial number of the machine in upper case. |

### Options

| Option | Default | Description |
| ------ | ------- | ----------- |
| `-sabakan-address` | (required) | Address of sabakan's GraphQL API, in the form `host:port`. |
| `-output` | `/var/lib/resolve-map/resolve-map.json` | Path of the resolve map file to write. The directory must exist. |
| `-interval` | `1m` | Interval to update the file. If `0`, it updates the file once and exits. |

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/bmc-reverse-proxy)
