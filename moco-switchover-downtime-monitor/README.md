# moco-switchover-downtime-monitor

**moco-switchover-downtime-monitor** is a diagnostic tool to monitor MySQLCluster downtime during [moco]'s switchover-like operations. This program executes the switchover-like operations and monitor downtime continuously. You should not monitor production clusters.

## Algorithm

This program runs the following procedure continuously.

1. Spawn goroutines to check endpoint availability.
   - The following endpoints are checked concurrently.
     - `all`, the Service points all MySQL instances. (i.e. no suffix)
     - `primary`, the Service points the primary instance. (i.e. `-primary` suffix)
     - `replica`, the Service points the replica instances. (i.e. `-replica` suffix)
   - For `all` and `replica` endpoints, a `SELECT` statement is executed. For `primary` endpoint, a `SELECT` statement and an `UPDATE` statement are executed.
2. Run a switchover-like operation on the target cluster.
   - The operation is one of the following:
     - `switchover`: run switchover by `kubectl moco switchover`.
     - `rollout`: run rolling update of the StatefulSet by `kubectl rollout restart`.
     - `poddelete`: delete a Pod by `kubectl delete pod`.
     - `killproc`: kill `mysqld` process by `kubectl exec -- kill`.
   - In case of `poddelete` and `killproc`, the target Pod is `primary` or `replica`.
3. Wait for the target cluster to back to be stable.
4. Report downtime as metrics.
5. Wait for a moment as a cool down.

## Gross and net downtime

The gross downtime is the duration from the first unavailability to the last unavailability. The net downtime is the sum of the unavailable duration.

For example, if an endpoint is first unavailable for 10 seconds and then available for 15 seconds and finally unavailable for 5 seconds, the gross downtime is 30 seconds (10+15+5) and the net downtime is 15 seconds (10+5).

## Invocation

```
moco-switchover-downtime-monitor [--namespace NAMESPACE] <cluster-name>...
```

- `--namespace`: The namespace in which the target clusters reside. If not specified, this program reads namespace information from `/var/run/secrets/kubernetes.io/serviceaccount/namespace`. i.e. if this program run in a Kubernetes Pod and the target clusters reside in the same namespace as this program, `--namespace` option is not required.
- `cluster-name`: The target MySQLCluster names to be checked. All clusters are checked concurrently.

## RBAC

This program requires the role described in [role.yaml](role.yaml).

## Database setup

The target MySQLClusters must be first initialized as written in [init.sql](init.sql). This program does not initialize the databases automatically.

## Metrics

The following metrics are exposed.

| Name                                                    | Description    | Labels                              | Type      | Value                     |
| ------------------------------------------------------- | -------------- | ----------------------------------- | --------- | ------------------------- |
| moco_switchover_downtime_monitor_downtime_gross_seconds | Gross downtime | cluster, operation, endpoint, write | Histogram | Gross downtime in seconds |
| moco_switchover_downtime_monitor_downtime_net_seconds   | Net downtime   | cluster, operation, endpoint, write | Histogram | Net downtime in seconds   |
| moco_switchover_downtime_monitor_check_failure_total    | Failure count  | cluster, operation, reason          | Counter   | The number of failures    |

- `operation` = `switchover` | `rollout` | `poddelete-primary` | `poddelete-replica` | `killproc-primary` | `killproc-replica`
- `endpoint`,`write` = `all`,`false` | `primary`,`true` | `primary`,`false` | `replica`,`false`
- `reason` = `execution_timeout`, `execution_failure`, `completion_timeout`, `completion_failure`

[moco]: https://github.com/cybozu-go/moco
