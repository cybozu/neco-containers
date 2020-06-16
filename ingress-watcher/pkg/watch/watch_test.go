package watch

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

const timeoutDuration = 550 * time.Millisecond

const (
	httpGetSuccessfulTotalName  = "ingresswatcher_http_get_successful_total"
	httpsGetSuccessfulTotalName = "ingresswatcher_https_get_successful_total"
	httpGetFailTotalName        = "ingresswatcher_http_get_fail_total"
	httpsGetFailTotalName       = "ingresswatcher_https_get_fail_total"
)

var metricsNames = []string{
	httpGetSuccessfulTotalName,
	httpsGetSuccessfulTotalName,
	httpGetFailTotalName,
	httpsGetFailTotalName,
}

func TestWatcherRun(t *testing.T) {
	type fields struct {
		targetAddrs []string
		interval    time.Duration
		httpClient  *http.Client
	}
	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name   string
		fields fields
		result map[string]float64
	}{
		{
			name: "GET success every 100ms in 550ms",
			fields: fields{
				targetAddrs: []string{"foo", "bar"},
				interval:    100 * time.Millisecond,
				httpClient: newTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
						Header:     make(http.Header),
					}, nil
				}),
			},
			result: map[string]float64{
				httpGetSuccessfulTotalName:  5,
				httpsGetSuccessfulTotalName: 5,
				httpGetFailTotalName:        0,
				httpsGetFailTotalName:       0,
			},
		},

		{
			name: "GET fail every 100ms in 550ms",
			fields: fields{
				targetAddrs: []string{"foo"},
				interval:    100 * time.Millisecond,
				httpClient: newTestClient(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				}),
			},
			result: map[string]float64{
				httpGetSuccessfulTotalName:  0,
				httpsGetSuccessfulTotalName: 0,
				httpGetFailTotalName:        5,
				httpsGetFailTotalName:       5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			metrics.HTTPGetSuccessfulTotal.Reset()
			metrics.HTTPSGetSuccessfulTotal.Reset()
			metrics.HTTPGetFailTotal.Reset()
			metrics.HTTPSGetFailTotal.Reset()
			registry.MustRegister(
				metrics.HTTPGetSuccessfulTotal,
				metrics.HTTPGetFailTotal,
				metrics.HTTPSGetSuccessfulTotal,
				metrics.HTTPSGetFailTotal,
			)

			// create watcher and run
			w := NewWatcher(
				tt.fields.targetAddrs,
				tt.fields.interval,
				tt.fields.httpClient,
			)
			env := well.NewEnvironment(context.Background())
			env.Go(func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
				defer cancel()
				return w.Run(ctx)
			})
			env.Stop()
			env.Wait()

			// parse mertics family
			metricsFamily, err := registry.Gather()
			if err != nil {
				t.Fatal(err)
			}

			type metricKey struct {
				name string
				path string
			}
			mfMap := make(map[metricKey]*dto.Metric)
			for _, mf := range metricsFamily {
				if len(mf.Metric) != len(tt.fields.targetAddrs) {
					t.Fatalf("%s: metric %s should contain only one element.", tt.name, *mf.Name)
				}
				for _, met := range mf.Metric {
					p := labelToMap(met.Label)["path"]
					mfMap[metricKey{*mf.Name, p}] = met
				}
			}

			// assert results
			for _, n := range metricsNames {
				for _, ta := range w.tagetAddrs {
					m, ok := mfMap[metricKey{n, ta}]
					if !ok && tt.result[n] != 0 {
						t.Errorf(
							"%s: value for %s{path=%s} should be %f but not found.",
							tt.name,
							n,
							ta,
							tt.result[n],
						)
						continue
					}
					if !ok && tt.result[n] == 0 {
						continue
					}

					v, ok := tt.result[n]
					if !ok {
						t.Fatalf("%s: value for %s{path=%s not found", tt.name, n, ta)
					}
					if v != *m.Counter.Value {
						t.Errorf(
							"%s: value for %s{path=%s} is wrong.  expected: %f, actual: %f",
							tt.name,
							n,
							ta,
							v,
							*m.Counter.Value,
						)
					}
				}
			}
		})
	}
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func labelToMap(labelPair []*dto.LabelPair) map[string]string {
	res := make(map[string]string)
	for _, l := range labelPair {
		res[*l.Name] = *l.Value
	}
	return res
}
