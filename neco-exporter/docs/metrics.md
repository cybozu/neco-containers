# Metrics

neco-exporter exports metrics from the following collectors.  
Each collector's name should appear in `--collectors` (comma-separated list) to activate.  
Each collector's scope should match `--scope` to use.

| Collector                 | Scope     | Description                         |
| ------------------------- | --------- | ----------------------------------- |
| [`bpf`](#bpf)             | `node`    | Measure BPF Program performance     |
| [`ciliumid`](#ciliumid)   | `cluster` | Count CiliumIdentity resources      |
| [`collector`](#collector) | (both)    | neco-exporter and collectors status |

## bpf

### `bpf_run_time_seconds_total`

Cumulative execution time for the BPF Program.

| Label       | Condition | Description                  |
| ----------- | --------- | ---------------------------- |
| `id`        | (Always)  | BPF Program ID               |
| `type`      | (Always)  | BPF Program Type             |
| `name`      | (Always)  | BPF Program Name             |
| `ifindex`   | TCX       | Attached TCX ifindex         |
| `direction` | TCX       | Attached TCX Direction       |
| `namespace` | TCX & Pod | Namespace of attached device |
| `pod`       | TCX & Pod | Pod of attached device       |
| `container` | TCX & Pod | Container of attached device |

### `bpf_run_count_total`

Execution count for the BPF Program.

See [`bpf_run_time_seconds_total`](#bpf_run_time_seconds_total) for the associated labels.

## ciliumid

### `ciliumid_identity_count`

Number of `CiliumIdentity` resources for the namespace.

| Label       | Description           |
| ----------- | --------------------- |
| `namespace` | Namespace of Identity |

## collector

### `collector_leader`

Report if the pod is elected as a leader by controller-runtime.

### `collector_health`

Report if the collector successfully collect its metrics or not.

| Label       | Description    |
| ----------- | -------------- |
| `collector` | Collector Name |

### `collector_process_seconds`

Elapsed time to collect metrics from the collector.

| Label       | Description    |
| ----------- | -------------- |
| `collector` | Collector Name |
