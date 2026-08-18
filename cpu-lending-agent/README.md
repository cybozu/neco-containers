cpu-lending-agent
=================

`cpu-lending-agent` is an NRI plugin that lends the exclusive CPUs of idle
MySQL replicas to best-effort workloads, and takes them back when the replica
is promoted to primary.

## Background

MOCO MySQL Pods run with Guaranteed QoS and integer CPU requests, so the
kubelet static CPU manager gives the `mysqld` container exclusive CPUs.
This is required for primary instances, but replicas use only a few percent
of their reserved CPUs. On nodes shared with other workloads, these idle
CPUs are wasted.

`cpu-lending-agent` recovers this capacity without touching the kubelet,
CKE, or MOCO:

- While a MySQL Pod has the label `moco.cybozu.com/role=replica`, its
  exclusive CPUs are added to the cpuset of opted-in BestEffort Pods
  ("borrowers") on the same node.
- When the Pod is promoted (`moco.cybozu.com/role=primary`), the agent
  removes those CPUs from all borrowers. This is a single cgroup update per
  borrower and completes in tens of milliseconds. Full exclusivity is
  restored to the same state as a node without the agent.
- While lending is active, the CPU scheduler still favors MySQL: a
  BestEffort borrower competes against the Guaranteed MySQL Pod with a
  cgroup weight ratio of about 1:80 or less, so MySQL gets 99%+ of a lent
  CPU whenever it wants it, even if the agent is down.

## How it works

The agent runs as a DaemonSet, one Pod per node, and connects to:

1. **containerd via NRI** (`/var/run/nri/nri.sock`): it rewrites the cpuset
   in `CreateContainer` and `UpdateContainer` requests for borrower
   containers, and issues unsolicited container updates to reclaim CPUs.
   Because every kubelet-originated cpuset write passes through the NRI
   hooks, there is no write conflict with the kubelet CPU manager.
2. **kube-apiserver**: an informer watches the Pods on its own node
   (filtered by `spec.nodeName`) for role label changes. Read-only.
3. **`/var/lib/kubelet/cpu_manager_state`** (read-only): the source of truth
   for exclusive CPU assignments and the shared pool.

The reconciliation is level-triggered: on every event the agent recomputes
the desired cpuset of all borrowers from scratch and updates only the
containers that differ. If the agent restarts, the NRI `Synchronize` hook
provides the full container list to converge from.

The agent never writes to the `mysqld` container's cgroup, the Kubernetes
API, or the kubelet state. All failure modes converge to the current
production behavior (full pinning, no lending).

## Requirements

- kubelet with `cpuManagerPolicy: static` (the agent goes inert and raises
  an alert if the policy is not static)
- containerd 2.0+ with NRI enabled (default)
- cgroup v2

## Usage

A Pod becomes a borrower by requesting the extended resource — the request
is the single opt-in signal, there is no label. The resource is denominated
in **milli-CPUs** (1000 = one lent CPU), following the same convention as
Koordinator's `kubernetes.io/batch-cpu`:

```yaml
resources:
  requests:
    cpu-lending.cybozu.io/preemptible-millicpu: "1000"
  limits:
    cpu-lending.cybozu.io/preemptible-millicpu: "1000"
```

Never use the `m` suffix. Quantity suffixes scale the number itself, not
the unit: `"1000m"` parses as `1` (a thousandth of what you meant) and is
accepted silently, while `"500m"` (= 0.5) is rejected because extended
resources must be integers. Write plain integers (`"1000"`, `"500"`) or the
`k` suffix (`"1k"` = 1000, matching how `kubectl describe node` displays the
values). The value controls only how many borrowers the scheduler packs per
node; at runtime every borrower on the node shares the whole lent CPU set
under kernel weight arbitration.

Borrower Pods must accept the following contract:

- **BestEffort QoS**: no cpu/memory requests or limits at all (extended
  resources do not affect the QoS class). The agent refuses to lend to Pods
  of other QoS classes because the scheduling weight guarantee does not hold
  for them. Note that even a memory limit alone makes a Pod Burstable.
- **CPUs can be taken away at any time** (on primary promotion), with no
  notice. This is a cgroup update; the process keeps running, just slower.
- **First to die under memory pressure**: with no memory reservation, the
  borrower is the first target of both kubelet eviction and the OOM killer.

Namespace-level governance is available with a standard ResourceQuota on
`requests.cpu-lending.cybozu.io/preemptible-millicpu`. To catch the `m`-suffix
mistake at admission time, deploy the ValidatingAdmissionPolicy in
`e2e/vap.yaml`, which rejects requests below 100 milli-CPUs with an
explanatory message.

Lender Pods are selected by the MOCO Pod labels
(`app.kubernetes.io/created-by=moco`, `app.kubernetes.io/name=mysql`) and
lending follows the `moco.cybozu.com/role` label. A missing role label
(e.g. during failover) means "do not lend" (fail-closed).

## Lending capacity as an extended resource

The agent advertises the current lending capacity of its node as the same
extended resource in the Node status. With the request above, the scheduler
places borrowers only onto nodes with lendable CPUs and bounds the number of
borrowers per node.

For humans, the agent also maintains a node annotation with the arithmetic
already done — read this instead of the Capacity/Allocated tables (those are
the scheduler's internal ledger and are displayed in canonicalized units
like `2k`):

```
kubectl describe node <node>   # in the Annotations section:
cpu-lending.cybozu.io/status:
  capacity_milli=2000 allocated_milli=1000 free_milli=1000
```

A negative `free_milli` is the expected over-allocated state right after a
promotion (capacity shrank while running borrowers keep their requests); it
resolves as the borrowers finish.

The ledger follows reality: on promotion the agent reclaims the cpusets
first and then shrinks the capacity. Shrinking does not evict running
borrowers; they keep running on the shared pool and the over-allocated
ledger resolves when they finish. New placements are blocked immediately.

This requires `get` on `nodes` and `patch` on `nodes/status` in addition to
the read-only Pod access. The resource name is configurable with
`-preemptible-millicpu-resource`.

## Docker images

Docker images are available on
[ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/cpu-lending-agent)
