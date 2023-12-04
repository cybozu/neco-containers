bpf-map-pressure-exporter
===================

`bpf-map-pressure-exporter` exposes BPF map pressure.

## Config
Default config file is `/etc/bpf-map-pressure-exporter/config.yaml`.
The target BPF maps should be specified under `mapNames`.

```yaml
mapNames:
  - cilium_ct
  - ...
fetchInterval: 30s
```

`mapNames` are interpreted as substrings of the map names and the pressure metrics for all maps including the substring are exposed.
If multiple `mapNames` are specified and some of them match the same map, only 1 metric is exposed.
Note that BPF map names are truncated to 15 characters.

`fetchInterval` defines the time interval to fetch BPF map pressure.
`bpf-map-pressure-exporter` fetches the map pressure of target maps every `fetchInterval` and returns the latest value when scraped.
Default value is `30s`.

## Usage
Command-line options are:

| Option         | Default value                                   | Description                          |
| -------------- | ----------------------------------------------- | ------------------------------------ |
| `port`         | `8080`                                          | port number to export metrics        |
| `config`       | `/etc/bpf-map-pressure-exporter/config.yaml`    | config file path                     |

API endpoints are:

| Path     | Description                 |
| -------- | --------------------------- |
| /health  | the path for liveness probe |
| /metrics | exporting metrics           |

## Prometheus metrics

`bpf-map-pressure-exporter` exposes the following metrics.

### `bpf_map_pressure`

`bpf_map_pressure` is a gauge that indicates the BPF map pressure.

| Label      | Description                |
| ---------- | -------------------------- |
| `map_id`   | ID of the BPF map          |
| `map_name` | name of the BPF map        |
