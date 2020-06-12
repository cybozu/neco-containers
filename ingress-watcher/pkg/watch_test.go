package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/ingress-watcher/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const timeoutDuration = 550 * time.Millisecond

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

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
		result int
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
			result: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := prometheus.NewRegistry()
			registry.MustRegister(metrics.HTTPGetSuccessfulTotal, metrics.HTTPSGetSuccessfulTotal)

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

			metrics, _ := registry.Gather()
			fmt.Printf("%#v", metrics)
		})
	}
}
