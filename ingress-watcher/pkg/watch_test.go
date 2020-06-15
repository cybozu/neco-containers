package pkg

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const timeoutDuration = 550 * time.Millisecond

const (
	httpGetSuccessfulTotalName  = "ingresswatcher_http_get_successful_total"
	httpsGetSuccessfulTotalName = "ingresswatcher_https_get_successful_total"
	httpGetFailTotalName        = "ingresswatcher_http_get_fail_total"
	httpsGetFailTotalName       = "ingresswatcher_https_get_fail_total"
)

func TestWatcherRun(t *testing.T) {
	type fields struct {
		endpoint   string
		interval   time.Duration
		httpClient *http.Client
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
			name: "GET every 100ms in 550ms",
			fields: fields{
				endpoint: "foo",
				interval: 100 * time.Millisecond,
				httpClient: newTestClient(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(""))),
						Header:     make(http.Header),
					}
				}),
			},
			result: map[string]float64{
				httpGetSuccessfulTotalName:  5,
				httpsGetSuccessfulTotalName: 5,
				httpGetFailTotalName:        0,
				httpsGetFailTotalName:       0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			registry.MustRegister(
				metrics.HTTPGetSuccessfulTotal,
				metrics.HTTPGetFailTotal,
				metrics.HTTPSGetSuccessfulTotal,
				metrics.HTTPSGetFailTotal,
			)

			w := &Watcher{
				endpoint:   tt.fields.endpoint,
				interval:   tt.fields.interval,
				httpClient: tt.fields.httpClient,
			}

			env := well.NewEnvironment(context.Background())
			env.Go(func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
				defer cancel()
				return w.Run(ctx)
			})
			env.Stop()
			env.Wait()

			metricsFamily, err := registry.Gather()
			if err != nil {
				t.Fatal(err)
			}

			for _, mf := range metricsFamily {
				if mf.Name == nil {
					t.Fatalf("%s: name should no be nil", tt.name)
				}
				for _, m := range mf.Metric {
					v, ok := tt.result[*mf.Name]
					if !ok {
						t.Errorf("%s: value for %q is not found", tt.name, *mf.Name)
					}
					if v != *m.Counter.Value {
						t.Errorf(
							"%s: value for %q is wrong.  expected: %f, actual: %f",
							tt.name,
							*mf.Name,
							*m.Counter.Value,
							v,
						)
					}
				}
			}
		})
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
