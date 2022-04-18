ceph-extra-exporter
===================

`ceph-extra-exporter` exposes extra metrics using some Ceph commands.

## Prometheus metrics

`ceph-extra-exporter` exposes the following metrics.

### `ceph_extra_build_info`

`ceph_extra_build_info` is a gauge that indicates the version number.

| Label     | Description                |
| --------- | -------------------------- |
| `version` | version number as a string |

### `ceph_extra_failed_total`

`ceph_extra_failed_total` is a counter that indicates error counts with a digest of reason.

| Label       | Description               |
| ----------- | ------------------------- |
| `reason`    | the reason of error       |
| `subsystem` | subsystem name of metrics |

### `ceph_extra_osd_pool_autoscale_status_pool_count`

`ceph_extra_osd_pool_autoscale_status_pool_count` is a counter that indicates the number of pools.

| Label | Description |
| ----- | ----------- |
| none  | none        |

## How to add a metrics

Add a new rule to `main.go` like below.

```go
var rules = []rule{
    <existing rules...>
    {
        name:    "<metrics subsystem name. it is usually the command and options joined by `-`.>",
        command: []string{"<the command output json formatted text.>"},
        metrics: map[string]metric{
            "<metrics name.>": {
                metricType: <metrics type e.g. `prometheus.GaugeValue`>,
                help:       "<help string.>",
                jqFilter:   "<the jq command's filter to get a single number as a metrics.>",
            },
            <can get some metrics from a command.>
        },
    }
```
