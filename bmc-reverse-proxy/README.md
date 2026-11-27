bmc-reverse-proxy-container
============================

`bmc-reverse-proxy` is a reverse proxy server for BMC.

The following products are assumed as BMC.

- dell iDRAC.

This program provides reverse proxy at once for the two ports (443, 5900) used by iDRAC.

It resolves the BMC address from the serial of the machine in the host name of the access destination
(e.g. `abc1234.bmc.example.com`), using the resolve map file.

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
3. Edit `-sabakan-url` and the domain of the `Certificate` (`*.bmc.example.com`) in `bmc-reverse-proxy.yaml`, and deploy bmc-reverse-proxy.
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

   curl -k --resolve <SERIAL>.bmc.example.com:443:<EXTERNAL-IP> https://<SERIAL>.bmc.example.com
   ```

   The BMCs of placemat do not provide Web interfaces, so the request fails with 502.
   Check that the `proxy error` log of `bmc-reverse-proxy` shows the BMC address of the machine in `backend`.
   If the request fails with 404 and `unknown host` is logged instead, the serial is not in the resolve map.

   ```console
   kubectl logs deploy/bmc-reverse-proxy -c bmc-reverse-proxy
   ```

Options
-------

### `bmc-reverse-proxy`

| Option         | Default                                 | Description                                                   |
| -------------- | --------------------------------------- | ------------------------------------------------------------- |
| `-tls-cert`    | `/etc/tls/tls.crt`                      | Path of the TLS certificate file.                             |
| `-tls-key`     | `/etc/tls/tls.key`                      | Path of the TLS private key file.                             |
| `-resolve-map` | `/var/lib/resolve-map/resolve-map.json` | Path of the resolve map file written by `update-resolve-map`. |

`bmc-reverse-proxy` re-reads the TLS certificate and the resolve map file every minute.
Until the resolve map file is loaded for the first time, it retries reading the file every second.
If the resolve map file cannot be read, it logs an error and keeps using the last loaded map.
While waiting for the file to be written for the first time, it logs `waiting for resolve map` instead of an error.
If the TLS certificate cannot be read at startup, it exits.
After that, it logs an error and keeps using the current certificate, so that active sessions are not cut.

Requests for a hostname that is not in the resolve map are not forwarded.
`bmc-reverse-proxy` returns 404 and logs `unknown host` at WARN level,
so that they can be distinguished from failures of BMCs (502).
Requests canceled by clients, e.g. by closing the browser tab, are recorded with 499 in the access log
(a convention of nginx) without an error log.

### Probes

`bmc-reverse-proxy` serves the following endpoints for the probes of Kubernetes on port 8080.

| Path | Description |
| ---- | ----------- |
| `/healthz` | Returns 200 while the process is running. For liveness probe. |
| `/readyz` | Returns 200 after the resolve map file and the TLS certificate have been loaded, and 503 before that. For readiness probe. |

Until they are loaded, the Pod is not ready and receives no traffic via the Service.
Once loaded, it stays ready even if the file cannot be read later, because the last loaded map is kept.

### Local testing

`bmc-reverse-proxy` always serves HTTPS, because the Web interfaces of BMCs redirect clients to `https://` URLs.
When you run it on your machine to check its behavior, use a self-signed certificate.

1. Create a self-signed certificate.

   ```console
   openssl req -x509 -newkey rsa:2048 -nodes -days 30 -subj "/CN=example-host.localhost" \
     -addext "subjectAltName=DNS:example-host.localhost" \
     -keyout /tmp/tls.key -out /tmp/tls.crt
   ```

2. Write a resolve map file, which maps names to addresses. In production, the names are serials of machines.
   The keys are case-insensitive.
   The address should be a server that listens on HTTPS (443 for port 8443, 5900 for port 5900).
   In the following steps, you can use any name instead of `example-host`.

   ```console
   echo '{"example-host": "<ADDRESS>"}' > /tmp/resolve-map.json
   ```

3. Run `bmc-reverse-proxy` with the certificate and the resolve map file.

   ```console
   go run ./cmd/bmc-reverse-proxy -tls-cert /tmp/tls.crt -tls-key /tmp/tls.key -resolve-map /tmp/resolve-map.json
   ```

4. Access `https://<HOSTNAME>.localhost:8443/` with a browser or curl.
   The first label of the hostname is used to look up the resolve map,
   so you cannot use `localhost` or an IP address as is.
   The browser shows a warning because the certificate is self-signed. Accept it to continue.

   ```console
   curl -k https://example-host.localhost:8443/
   ```

   Major browsers (Chrome, Firefox, Edge) and curl resolve names ending with `.localhost` to the loopback address by themselves,
   so you do not need to edit `/etc/hosts`.
   Note that other tools using the OS resolver may not resolve them.

   To access the proxy from another machine, add the hostname to `/etc/hosts` of that machine instead,
   e.g. `192.0.2.10 example-host.bmc.local`, and access `https://example-host.bmc.local:8443/`.

You can also write the file with `update-resolve-map` if sabakan is reachable from your machine.

```console
go run ./cmd/update-resolve-map -sabakan-url http://<SABAKAN_ADDRESS>:10080/ -output /tmp/resolve-map.json -interval 0
```

update-resolve-map
------------------

`update-resolve-map` writes the resolve map file, which maps serials of machines to BMC addresses.
It gets the machines from sabakan's REST API (`GET /api/v1/machines`) every `-interval`, and writes the file only when the content is changed.
It logs `updated resolve map` when it writes the file.
The file is replaced atomically, so `bmc-reverse-proxy` never reads a partially written file.
If sabakan cannot be reached, or no BMC addresses are found in the response of sabakan,
it logs an error and keeps the last written file.
Until the file is written for the first time, it retries every 5 seconds instead of every `-interval`.

The file is a JSON object like below.

```json
{
  "ABC1234": "192.0.2.4",
  "XYZ5678": "192.0.2.5"
}
```

The keys are the serials of the machines in upper case, as shown by sabakan and BMCs.
Machines without a BMC address are skipped.
Retired machines are also included because we need to operate them via BMCs.

`bmc-reverse-proxy` takes the first label of the host of a request, e.g. `abc1234` in `abc1234.bmc.example.com`,
and looks it up in the file case-insensitively, because hostnames are case-insensitive.

### Options

| Option | Default | Description |
| ------ | ------- | ----------- |
| `-sabakan-url` | `http://localhost:10080/` | URL of sabakan's REST API. |
| `-output` | `/var/lib/resolve-map/resolve-map.json` | Path of the resolve map file to write. The directory must exist. |
| `-interval` | `1m` | Interval to update the file. If `0`, it updates the file once and exits. |

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/bmc-reverse-proxy)
