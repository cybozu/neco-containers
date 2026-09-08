# Metrics

neco-exporter exports metrics from the following collectors.  
Each collector's name should appear in `--collectors` (comma-separated list) to activate.  
Each collector's scope should match `--scope` to use.

| Collector                 | Scope     | Description                         |
| ------------------------- | --------- | ----------------------------------- |
| [`bpf`](#bpf)                         | `node`    | Measure BPF Program performance     |
| [`cert`](#cert)                       | `cluster` | Monitor TLS certificate expiration  |
| [`ciliumid`](#ciliumid)               | `cluster` | Count CiliumIdentity resources      |
| [`kubelet`](#kubelet)                 | `node`    | Report kubelet's systemReserved cpu/memory |
| [`networkfence`](#networkfence)       | `cluster` | Monitor NetworkFence resources      |
| [`nicirq`](#nicirq)                   | `node`    | Report which CPU handles each NIC queue interrupt |
| [`collector`](#collector)             | (both)    | neco-exporter and collectors status |

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

## cert

### `cert_expiration_time_seconds`

Expiration (`notAfter`) of TLS certificates.
It appears only for Kubernetes Secrets not maintained by cert-manager.

| Label       | Description         |
| ----------- | ------------------- |
| `namespace` | Namespace of Secret |
| `name`      | Name of Secret      |

## ciliumid

### `ciliumid_identity_count`

Number of `CiliumIdentity` resources for the namespace.

| Label       | Description           |
| ----------- | --------------------- |
| `namespace` | Namespace of Identity |

## kubelet

### `kubelet_system_reserved`

CPU and memory reserved via kubelet's `systemReserved` config (as opposed to
`kubeReserved`, which this collector does not report), read once at startup.
Labeled the same way as `kube_node_status_allocatable`.

| Label      | Description                                  |
| ---------- | --------------------------------------------- |
| `node`     | Node name (from the `NODE_NAME` env var)      |
| `resource` | Reserved resource (`cpu` or `memory`)         |
| `unit`     | Unit of `resource` (`core` or `byte`)         |

## nicirq

### `nicirq_queue_cpu`

The CPU that a NIC queue interrupt is actually delivered to, read from
`/proc/irq/<N>/effective_affinity_list` on every collection. One series per
queue interrupt whose action name looks like `<driver>-<device>-TxRx-<queue>`
(the naming used by the `ice` driver).

This exists to detect the placement that degraded MySQL in 2026-04: after the
`ice` driver stopped applying IRQ affinity itself, the kernel default can put
every queue interrupt of a NIC on the CPUs of one NUMA node. The spread is
derived in PromQL, per device, as the number of distinct CPUs divided by the
number of queues; 1.0 means one CPU per queue, 0.125 means 64 queues on 8 CPUs.

```
count by (node, device) (count_values by (node, device) ("cpu", neco_node_nicirq_queue_cpu))
  / count by (node, device) (neco_node_nicirq_queue_cpu)
```

`smp_affinity` and `affinity_hint` are deliberately not exported: on
single-NUMA nodes the hint never matches the effective CPU, so they cannot
serve as a health signal.

| Label    | Description                                                        |
| -------- | ------------------------------------------------------------------ |
| `node`   | Node name (from the `NODE_NAME` env var)                           |
| `driver` | Driver prefix of the action name (`ice`)                           |
| `device` | Device part of the action name (`eno12399np0`, or `0000:c4:00.0:ctrl` for the control queue) |
| `queue`  | Queue index (`TxRx-<queue>`)                                       |

### `nicirq_skipped_queues`

Number of queue interrupts left out of `nicirq_queue_cpu` in the latest
collection, per reason. A queue that cannot be attributed to exactly one CPU is
skipped rather than failing the whole collection, so that one odd or vanishing
IRQ does not hide the metrics of the healthy ones. Every reason is reported on
every collection (0 when nothing was skipped); a persistent non-zero value means
the procfs layout or the driver's handler naming differs from what the collector
expects. A log line is written only when the counts change, not on every
collection; the metric is the signal to alert on. A node whose NIC stopped producing `nicirq_queue_cpu` altogether can be
found by joining on the scrape target (`collector_health` carries no `node`
label, but both series come from the same pod):

```
neco_node_collector_health{collector="nicirq"} == 1
  unless on (instance) count by (instance) (neco_node_nicirq_queue_cpu)
```

| Label    | Description                                                                 |
| -------- | --------------------------------------------------------------------------- |
| `node`   | Node name (from the `NODE_NAME` env var)                                    |
| `reason` | `read_error` (directory or `effective_affinity_list` unreadable), `not_single_cpu` (`effective_affinity_list` is not a single CPU number; on x86 a vector is bound to one CPU, so this covers an empty list, a range, or garbage), `unmatched_action` (contains `TxRx` but does not fit `<driver>-<device>-TxRx-<queue>`) |

## networkfence

### `networkfence_info`

Information about `NetworkFence` resources.

| Label         | Description                                  |
| ------------- | -------------------------------------------- |
| `name`        | Name of NetworkFence                         |
| `driver`      | CSI driver name                              |
| `fence_state` | Fence state (`Fenced`/`Unfenced`)            |
| `result`      | Operation result (e.g. `Succeeded`/`Failed`) |

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
