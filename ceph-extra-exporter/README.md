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

### `ceph_extra_rgw_bucket_stats_s3_object_count`

`ceph_extra_rgw_bucket_stats_s3_object_count` is a gauge metric that gives S3 Object count of RGW buckets.

| Label  | Description |
| ------ | ----------- |
| bucket | bucket name |

### `ceph_extra_rgw_bucket_stats_s3_size_bytes`

`ceph_extra_rgw_bucket_stats_s3_size_bytes` is a gauge metric that gives sum of S3 object size in the buckets.

| Label  | Description |
| ------ | ----------- |
| bucket | bucket name |

### `ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes`

`ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes` is a gauge metric that gives sum of S3 object size rounded to 4KBytes in the buckets.

| Label  | Description |
| ------ | ----------- |
| bucket | bucket name |

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
                jqFilter:   "<see jqFilter section.>",
                labelKeys: ["<a key for a label.>", ...]
            },
            <... can get multiple metrics from a command.>
        },
    }
```

### jqFilter

`jqFilter` must convert the output of the command to JSON in the following format and return it.
`value` is used as the value of the metric. The length of `labels` must be the same for all elements of the array and length of `labelKeys`.

```json
[
  {
    "value": "<metrics value>",
    "labels": [
      "label value",
      ...
    ]
  }
]
```

For an example of output...

```json
[
  {
    "value": 0,
    "labels": [
      "device_health_metrics"
    ]
  },
  {
    "value": 1.0802450726274402,
    "labels": [
      "ceph-ssd-block-pool"
    ]
  }
]
```
