package pkg

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

const timeoutDuration = 550 * time.Millisecond

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
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
			channel := make(chan string)
			w := &Watcher{
				endpoint:   tt.fields.endpoint,
				interval:   tt.fields.interval,
				channel:    channel,
				httpClient: tt.fields.httpClient,
			}
			well.Go(func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
				defer cancel()
				return w.Run(ctx)
			})
			well.Stop()

			buf := []string{}
			for v := range channel {
				buf = append(buf, v)
			}

			err := well.Wait()
			if err != nil {
				log.ErrorExit(err)
			}
			if len(buf) != tt.result {
				t.Errorf("%s Number of response: actual %d, expected %d", tt.name, len(buf), tt.result)
			}
		})
	}
}
