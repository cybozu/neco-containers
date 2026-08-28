package main

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/autoscale_status.json
var autoscale_status_json string

//go:embed testdata/bucket_stats.json
var bucket_stats_json string

//go:embed testdata/osd_df.json
var osd_df_json string

//go:embed testdata/rbd_task_list.json
var rbd_task_list_json string

func TestServer(t *testing.T) {
	testRules := rules
	testRules[0].command = []string{"echo", autoscale_status_json}
	testRules[1].command = []string{"echo", bucket_stats_json}
	testRules[2].command = []string{"echo", osd_df_json}
	testRules[3].command = []string{"echo", rbd_task_list_json}

	commonExpected := `# HELP ceph_extra_osd_df_crush_weight WEIGHT of ` + "`ceph osd df`" + ` command
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
`

	rgwExpected := `# HELP ceph_extra_rgw_bucket_stats_s3_object_count s3 object count of ` + "`radosgw-admin bucket stats`" + ` command
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
`
	rbdExpected := `# HELP ceph_extra_rbd_task_list_count RBD task count of ` + "`ceph rbd task list`" + ` command
# TYPE ceph_extra_rbd_task_list_count gauge
ceph_extra_rbd_task_list_count{action="flatten"} 1
ceph_extra_rbd_task_list_count{action="trash remove"} 2
`

	testcases := []struct {
		name        string
		port        uint
		options     exportOptions
		expected    string
		metricNames []string
		notContains []string
	}{
		{
			name:     "all metrics enabled",
			port:     8080,
			options:  exportOptions{rgwMetrics: true, rbdMetrics: true},
			expected: commonExpected + rgwExpected + rbdExpected,
			metricNames: []string{
				"ceph_extra_osd_pool_autoscale_status_pool_count",
				"ceph_extra_rgw_bucket_stats_s3_object_count",
				"ceph_extra_rgw_bucket_stats_s3_size_bytes",
				"ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes",
				"ceph_extra_osd_df_crush_weight",
				"ceph_extra_rbd_task_list_count",
			},
		},
		{
			name:     "only rgw metrics enabled",
			port:     8082,
			options:  exportOptions{rgwMetrics: true, rbdMetrics: false},
			expected: commonExpected + rgwExpected,
			metricNames: []string{
				"ceph_extra_osd_pool_autoscale_status_pool_count",
				"ceph_extra_rgw_bucket_stats_s3_object_count",
				"ceph_extra_rgw_bucket_stats_s3_size_bytes",
				"ceph_extra_rgw_bucket_stats_s3_size_rounded_bytes",
				"ceph_extra_osd_df_crush_weight",
			},
			notContains: []string{
				"ceph_extra_rbd_task_list_count",
			},
		},
		{
			name:     "only rbd metrics enabled",
			port:     8083,
			options:  exportOptions{rgwMetrics: false, rbdMetrics: true},
			expected: commonExpected + rbdExpected,
			metricNames: []string{
				"ceph_extra_osd_pool_autoscale_status_pool_count",
				"ceph_extra_osd_df_crush_weight",
				"ceph_extra_rbd_task_list_count",
			},
			notContains: []string{
				"ceph_extra_rgw_bucket_stats",
			},
		},
		{
			name:     "rgw and rbd metrics disabled",
			port:     8081,
			options:  exportOptions{rgwMetrics: false, rbdMetrics: false},
			expected: commonExpected,
			metricNames: []string{
				"ceph_extra_osd_pool_autoscale_status_pool_count",
				"ceph_extra_osd_df_crush_weight",
			},
			notContains: []string{
				"ceph_extra_rgw_bucket_stats",
				"ceph_extra_rbd_task_list_count",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testServerConfig(tc.port)
			cfg.options = tc.options
			go startServer(testRules, prometheus.NewRegistry(), cfg)
			url := fmt.Sprintf("http://localhost:%d/v1/metrics", tc.port)

			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				err := testutil.ScrapeAndCompare(
					url,
					strings.NewReader(tc.expected),
					tc.metricNames...,
				)
				assert.NoError(c, err)

				if len(tc.notContains) == 0 {
					return
				}

				_, body, err := get(tc.port, "/v1/metrics")
				require.NoError(c, err)
				for _, s := range tc.notContains {
					assert.NotContains(c, body, s)
				}
			}, 1*time.Minute, 100*time.Millisecond)
		})
	}
}

func getHealth(port uint) (int, string, error) {
	return get(port, "/v1/health")
}

// fastHealthConfig shrinks the timings so that a health test finishes in
// well under a second instead of minutes.
func fastHealthConfig(port uint) serverConfig {
	cfg := testServerConfig(port)
	cfg.executionInterval = 50 * time.Millisecond
	cfg.commandTimeout = 200 * time.Millisecond
	cfg.healthCheckThreshold = 500 * time.Millisecond
	return cfg
}

func get(port uint, path string) (int, string, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", port, path))
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, string(body), nil
}

// testServerConfig returns a serverConfig with the production defaults, which
// each test overrides only where it matters.
func testServerConfig(port uint) serverConfig {
	return serverConfig{
		port:                 port,
		options:              exportOptions{rgwMetrics: true, rbdMetrics: true},
		executionInterval:    executionInterval,
		commandTimeout:       commandTimeout,
		healthCheckThreshold: defaultHealthCheckThreshold,
	}
}

func healthTestRules(command []string) []rule {
	return []rule{
		{
			name:    "health_test",
			command: command,
			metrics: map[string]metric{
				"count": {
					metricType: prometheus.GaugeValue,
					help:       "test",
					jqFilter:   "[{value: . | length, labels: []}]",
				},
			},
		},
	}
}

// TestHealthWithFailingCommand confirms that a worker whose command keeps
// failing is still reported as healthy, because restarting the pod does not
// fix such a failure.
func TestHealthWithFailingCommand(t *testing.T) {
	t.Parallel()

	var port uint = 8086
	go startServer(healthTestRules([]string{"false"}), prometheus.NewRegistry(), fastHealthConfig(port))

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		status, _, err := getHealth(port)
		if !assert.NoError(c, err) {
			return
		}
		assert.Equal(c, http.StatusOK, status)
	}, 10*time.Second, 100*time.Millisecond)

	// The worker keeps failing for longer than healthCheckThreshold, but it
	// must never be reported as unhealthy.
	assert.Never(t, func() bool {
		status, _, err := getHealth(port)
		return err != nil || status != http.StatusOK
	}, time.Second, 100*time.Millisecond)
}

// TestHealthWithStuckWorker confirms that a worker stuck in update() is
// reported as unhealthy. The command spawns a background process which
// inherits stdout, so reading stdout does not finish even after the command
// itself has exited and its context has timed out.
func TestHealthWithStuckWorker(t *testing.T) {
	t.Parallel()

	var port uint = 8087
	go startServer(healthTestRules([]string{"sh", "-c", "sleep 5 & echo '[]'"}), prometheus.NewRegistry(), fastHealthConfig(port))

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		status, body, err := getHealth(port)
		if !assert.NoError(c, err) {
			return
		}
		assert.Equal(c, http.StatusServiceUnavailable, status)
		assert.Contains(c, body, "health_test")
	}, 10*time.Second, 100*time.Millisecond)
}

func TestStartServerConfigValidation(t *testing.T) {
	testcases := map[string]func(cfg *serverConfig){
		"executionInterval is not positive": func(cfg *serverConfig) { cfg.executionInterval = 0 },
		"commandTimeout is not positive":    func(cfg *serverConfig) { cfg.commandTimeout = 0 },
		"healthCheckThreshold is shorter than the minimum": func(cfg *serverConfig) {
			cfg.healthCheckThreshold = executionInterval
		},
	}

	for name, breakConfig := range testcases {
		t.Run(name, func(t *testing.T) {
			cfg := testServerConfig(8088)
			breakConfig(&cfg)
			err := startServer(healthTestRules([]string{"echo", "[]"}), prometheus.NewRegistry(), cfg)
			assert.Error(t, err)
		})
	}
}

// TestDefaultHealthCheckThreshold guards defaultHealthCheckThreshold against
// a rule which gains metrics: the default must never be rejected by validate().
func TestDefaultHealthCheckThreshold(t *testing.T) {
	cfg := testServerConfig(0)
	assert.NoError(t, cfg.validate(rules))
}
