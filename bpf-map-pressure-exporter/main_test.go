package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	fetcher := &mockBpfMapPressureFetcher{
		metricsFunc: func() []bpfMapPressure {
			return []bpfMapPressure{
				{
					mapId:       1,
					mapName:     "cilium_test_1",
					mapPressure: 0.1,
				},
				{
					mapId:       2,
					mapName:     "cilium_test_2",
					mapPressure: 0.2,
				},
			}
		},
		startFunc: func(ctx context.Context) {
		},
	}
	var port uint = 8080
	expect := `# HELP bpf_map_pressure bpf map pressure
# TYPE bpf_map_pressure gauge
bpf_map_pressure{map_id="1",map_name="cilium_test_1"} 0.1
bpf_map_pressure{map_id="2",map_name="cilium_test_2"} 0.2
`
	go startServer(fetcher, port)
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		err := testutil.ScrapeAndCompare(
			fmt.Sprintf("http://localhost:%d/metrics", port),
			strings.NewReader(expect),
			"bpf_map_pressure",
		)
		assert.NoError(c, err)
	}, 1*time.Minute, 5*time.Second)
}
