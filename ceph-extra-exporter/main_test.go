package main

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/autoscale_status.json
var autoscale_status_json string

//go:embed testdata/bucket_stats.json
var bucket_stats_json string

//go:embed testdata/osd_df.json
var osd_df_json string

func TestServer(t *testing.T) {
	testRules := rules
	testRules[0].command = []string{"echo", autoscale_status_json}
	testRules[1].command = []string{"echo", bucket_stats_json}
	testRules[2].command = []string{"echo", osd_df_json}

	var port uint = 8080
	go startServer(testRules, port, true)

	expected := strings.NewReader(`# HELP ceph_extra_osd_df_crush_weight WEIGHT of ` + "`ceph osd df`" + ` command
# TYPE ceph_extra_osd_df_crush_weight gauge
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.0"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.1"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.10"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.11"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.12"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.13"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.14"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.15"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.16"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.17"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.18"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.19"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.2"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.20"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.21"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.22"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.23"} 0.078094482421875
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.3"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.4"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.5"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.6"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.7"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.8"} 10.692398071289062
ceph_extra_osd_df_crush_weight{ceph_daemon="osd.9"} 0.078094482421875
# HELP ceph_extra_osd_pool_autoscale_status_pool_count pool count of ` + "`ceph osd pool autoscale-status`" + ` command
# TYPE ceph_extra_osd_pool_autoscale_status_pool_count gauge
ceph_extra_osd_pool_autoscale_status_pool_count 2
# HELP ceph_extra_rgw_bucket_stats_s3_object_count s3 object count of ` + "`radosgw-admin bucket stats`" + ` command
# TYPE ceph_extra_rgw_bucket_stats_s3_object_count gauge
ceph_extra_rgw_bucket_stats_s3_object_count{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"} 136473
ceph_extra_rgw_bucket_stats_s3_object_count{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"} 0
ceph_extra_rgw_bucket_stats_s3_object_count{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"} 550
# HELP ceph_extra_rgw_bucket_stats_s3_size_bytes sum of s3 objects bytes ` + "`radosgw-admin bucket stats`" + ` command
# TYPE ceph_extra_rgw_bucket_stats_s3_size_bytes gauge
ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"} 6.429367944e+09
ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"} 0
ceph_extra_rgw_bucket_stats_s3_size_bytes{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"} 180648
# HELP ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes sum of s3 objects bytes rounded to 4KBytes ` + "`radosgw-admin bucket stats`" + ` command
# TYPE ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes gauge
ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="loki-data-bucket-bedc5054-a90f-41b0-82f8-c077c2c32217"} 6.723739648e+09
ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="rook-ceph-bucket-checker-193e3cc1-063c-4d44-8a1a-cf147c682680"} 0
ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes{bucket="session-log-bucket-3d9a7583-f11b-4186-b4bc-8bf84c852662"} 2.326528e+06
`)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := testutil.ScrapeAndCompare(
			fmt.Sprintf("http://localhost:%d/v1/metrics", port),
			expected,
			"ceph_extra_osd_pool_autoscale_status_pool_count",
			"ceph_extra_rgw_bucket_stats_s3_object_count",
			"ceph_extra_rgw_bucket_stats_s3_size_bytes",
			"ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes",
			"ceph_extra_osd_df_crush_weight",
		)
		assert.NoError(c, err)
	}, 1*time.Minute, 5*time.Second)
}
