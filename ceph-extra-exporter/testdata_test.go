package main

import _ "embed"

//go:embed testdata/autoscale_status.json
var autoscale_status_json string

//go:embed testdata/bucket_stats.json
var bucket_stats_json string
