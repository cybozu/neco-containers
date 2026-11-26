# Metrics

neco-exporter exports metrics from the following collectors.  
Each collector's name should appear in `--collectors` (comma-separated list) to activate.  
Each collector's scope should match `--scope` to use.

| Collector                 | Scope     | Description                         |
| ------------------------- | --------- | ----------------------------------- |
| [`bpf`](#bpf)                         | `node`    | Measure BPF Program performance     |
| [`cert`](#cert)                       | `cluster` | Monitor TLS certificate expiration  |
| [`ciliumid`](#ciliumid)               | (both)    | Count and describe Cilium identities |
| [`cpupinning`](#cpupinning)           | `node`    | Report CPUs exclusively allocated to containers |
| [`kubelet`](#kubelet)                 | `node`    | Report kubelet's systemReserved cpu/memory |
| [`networkfence`](#networkfence)       | `cluster` | Monitor NetworkFence resources      |
| [`nicirq`](#nicirq)                   | `node`    | Report which CPU handles each NIC queue interrupt |
| [`route`](#route)                     | `node`    | Watch default routes of the node   |
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

The `cluster` scope reports `CiliumIdentity` resources, and the `node` scope reports identities of Pods on the node.

### `ciliumid_identity_count`

Scope: `cluster`

Number of `CiliumIdentity` resources for the namespace.

| Label       | Description           |
| ----------- | --------------------- |
| `namespace` | Namespace of Identity |

### `ciliumid_identity_info`

Scope: `cluster`

Info metric with a constant value of `1` for each `CiliumIdentity` resource.
It exposes the `security-labels` of the identity, i.e. the labels Cilium used to allocate the identity.

| Label                  | Condition  | Description                                                                 |
| ---------------------- | ---------- | --------------------------------------------------------------------------- |
| `identity`             | (Always)   | Numeric identity (name of `CiliumIdentity`)                                 |
| `uid`                  | (Always)   | UID of `CiliumIdentity`                                                     |
| `namespace`            | Namespaced | Value of `k8s:io.kubernetes.pod.namespace` security label                   |
| `security_labels`      | (Always)   | All security labels in the same format as Hubble metrics' `identity` context |
| `label_<source>_<key>` | (Dynamic)  | One label for each security label; the set differs between identities       |

`security_labels` is the sorted list of `<source>:<key>=<value>` (or `<source>:<key>` if the value is empty) joined by `,`.
This is identical to the value of `source` / `destination` labels of Hubble metrics with `sourceContext=identity` / `destinationContext=identity`,
so it can be used to join them, e.g.:

```promql
hubble_flows_processed_total
  * on (source) group_left (identity, namespace)
    label_replace(neco_cluster_ciliumid_identity_info, "source", "$1", "security_labels", "(.*)")
```

Security label keys (`<source>:<key>`) are converted to `label_*` names by prefixing `label_` and
replacing every character other than `[a-zA-Z0-9_]` with `_`.
For example, `k8s:identity.neco.cybozu.io/app` becomes `label_k8s_identity_neco_cybozu_io_app`.
`k8s:io.kubernetes.pod.namespace` is not converted because it is exposed as `namespace`.
If multiple keys are converted to the same name, the value of the label becomes `<CONFLICTED>`.
This cannot be an actual Kubernetes label value. The original values are still available in `security_labels`.

### `ciliumid_identity_pods`

Scope: `node`

Number of Pods on the node having the identity, as reported by the local cilium-agent.
When no Pod on the node has the identity, the series disappears instead of becoming `0`,
so `unless ... offset <duration>` can detect identities newly appearing on the node.

Endpoints that are not Pods (e.g. the host) and Pods with a reserved identity (e.g. `reserved:init`) are not counted.

| Label             | Description                                                                     |
| ----------------- | ------------------------------------------------------------------------------- |
| `node`            | Name of the node                                                                |
| `security_labels` | Security labels of the identity, identical to `security_labels` of `ciliumid_identity_info` |

## cpupinning

### `cpupinning_info`

Info metric with a constant value of `1` for each CPU exclusively allocated to a container by kubelet's static CPU manager policy.
It is read from `List` of kubelet's PodResources API (`/var/lib/kubelet/pod-resources/kubelet.sock`).
Containers in the shared pool have no series.

| Label       | Description                                        |
| ----------- | -------------------------------------------------- |
| `node`      | Node name (from the `NODE_NAME` env var)           |
| `namespace` | Namespace of the Pod                               |
| `pod`       | Name of the Pod                                    |
| `container` | Name of the container                              |
| `cpu`       | Logical CPU number, same as `cpu` of node_exporter |

Join with node_exporter to get the usage of the pinned CPUs, e.g. when its series have the `node` label:

```promql
sum without (mode) (rate(node_cpu_seconds_total{mode!~"idle|iowait"}[5m]))
  * on (node, cpu) group_left (namespace, pod, container)
    neco_node_cpupinning_info
```

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

## route

The `route` collector watches the IPv4 default route (`0.0.0.0/0`) in the main routing table of the node.
IPv6 routes are ignored.
It keeps a copy of the IPv4 main table, taken by dumps of the table.
It subscribes to route notifications of the kernel (rtnetlink), and dumps the table when they arrive:

- within 100 milliseconds after a notification of a default route;
- within 1 second after a notification of another route, so that frequent changes do not dump the table too often;
- every 30 seconds without notifications, because Linux does not notify some removals, e.g. routes removed when their device goes down.

The copy is not built from the notifications, because they are not versioned, and the ones received during a dump may be older than the dump.

The Pod must run in the host network namespace (`hostNetwork: true`); otherwise it watches the routes of the Pod.
No capability is needed.
`collector_health` is `0` until the first dump succeeds.

Counters are reset when the exporter restarts.

Run this collector alone in a dedicated DaemonSet (`--collectors=route`).
When the node loses its default route, it may also lose the route to the scraper, so scrapes fail until the routes come back.
`route_default_present` has no samples during that time,
but the logs below record the outage, and `route_changes_total` keeps counting in the process and reports the changes at the first scrape after it.
This works only if the process keeps running without the network, so do not enable together with it:

- collectors using the Kubernetes API, e.g. the cluster-scope collectors, whose leader election stops the process when the lease is lost;
- collectors whose `Setup` or resource usage may stop or restart the Pod,
  because a restart resets the counters.

It also logs the changes of default routes,
so that the time of an outage can be found in the logs after the routes come back:

| Level  | Message                     | Attributes                             |
| ------ | --------------------------- | -------------------------------------- |
| `INFO` | `default route changed`     | `op`, `metric`, `protocol`, `nexthops` |
| `WARN` | `default route is missing`  | `detected_by`, `not_before`            |
| `INFO` | `default route is restored` | `detected_by`, `not_before`            |

`default route changed` is logged for every notification of a default route, when it is received.
`op` is the same as in [`route_changes_total`](#route_changes_total).
`nexthops` lists the next hops in the notification, e.g. `via 10.69.1.193 ifindex 2`.
Changes without notifications, e.g. next hops marked dead when their devices go down, are not logged.

`default route is missing` and `default route is restored` are logged when the node loses all of its default routes and when one comes back.
They are found by dumps.
The change happened between `not_before` and the time of the log.

- `detected_by` is `notification` if a notification of a default route triggered the dump.
  `not_before` is the time of the first notification since the previous dump, so the change is logged within 100 milliseconds.
- `detected_by` is `dump` if the dump was periodic or triggered by the other routes, e.g. when the kernel removed the route without a notification.
  `not_before` is the time of the previous dump, up to 30 seconds earlier.

`default route is missing` is also logged at the start if the node has no default route, without `not_before`.

### `route_default_present`

`1` if the main table has any default route, `0` otherwise.

| Label    | Description                              |
| -------- | ---------------------------------------- |
| `node`   | Node name (from the `NODE_NAME` env var) |

### `route_default_nexthops`

Number of usable next hops of the default routes currently in the main table.
A multipath route has one next hop for each `nexthop` line of `ip route`.
Next hops marked `dead` or `linkdown` are not counted.
The kernel keeps them in multipath routes, without notifications, while their devices are down,
and keeps a single-path route with a `linkdown` next hop, whose value is `0`.
The series disappears while the route is missing.
It is updated by the dumps, like the other metrics of default routes.

| Label      | Description                                                 |
| ---------- | ----------------------------------------------------------- |
| `node`     | Node name (from the `NODE_NAME` env var)                    |
| `metric`   | Metric (priority) of the route, e.g. `32` for BIRD's routes |

### `route_changes_total`

Number of IPv4 route notifications for the main table.
A steady rate of `op="delete"` with `protocol="bird"` means that BIRD deletes and adds the same routes repeatedly.

| Label      | Description                                                                                  |
| ---------- | -------------------------------------------------------------------------------------------- |
| `node`     | Node name (from the `NODE_NAME` env var)                                                     |
| `protocol` | Routing protocol of the route                                                                |
| `op`       | `add`, `replace` (an existing route was replaced, e.g. its next hops changed), or `delete` |
| `default`  | `true` for default routes, `false` for the other routes                                     |

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
