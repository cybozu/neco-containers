# ct-monitor

[ct-monitor](https://github.com/Hsn723/ct-monitor) monitors [Certificate Transparency](https://certificate.transparency.dev/) logs via the [Cert Spotter](https://sslmate.com/certspotter/) API and sends email alerts when new certificate issuances are observed for configured domains.

## Plugins

Plugins are executables that implement the `IssuanceFilter` interface from `github.com/Hsn723/ct-monitor/filter`. They receive a list of issuances and return a filtered list. Plugins are placed under the `plugin/` directory and built into the container image at `/plugins/`.

### incluster-filter

Filters out issuances whose certificate fingerprint (SHA256 of the full certificate DER) matches a certificate already present in the cluster. Only CertificateRequests issued by the following ClusterIssuers are considered:

- `clouddns`
- `clouddns-letsencrypt`

The filter requires `get` and `list` permissions on `certificaterequests.cert-manager.io` across all namespaces.

## Docker image

The container image includes:

- `/ct-monitor` — the ct-monitor binary
- `/plugins/incluster-filter` — the incluster-filter plugin binary
