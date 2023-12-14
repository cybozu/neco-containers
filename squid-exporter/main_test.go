package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func serverFail(step string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch step {
		case "counters": // close connection to simulate failure
			if r.URL.Path == "/squid-internal-mgr/counters" {
				c, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					log.Fatal(err)
				}
				c.Close()
			} else {
				w.Write([]byte(""))
			}
		case "service_times": // close connection to simulate failure
			if r.URL.Path == "/squid-internal-mgr/service_times" {
				c, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					log.Fatal(err)
				}
				c.Close()

			} else {
				w.Write([]byte(""))
			}
		case "convert_counters": // return invalid data to simulate failure
			if r.URL.Path == "/squid-internal-mgr/counters" {
				w.Write([]byte(`sample_time = 1701938593.739082 (Thu, 07 Dec 2023 08:43:13 GMT)
					client_http.requests = x`))

			} else {
				w.Write([]byte(""))
			}
		case "convert_service_times": // return invalid data to simulate failure
			if r.URL.Path == "/squid-internal-mgr/service_times" {
				w.Write([]byte(`Service Time Percentiles            5 min    60 min:
					Cache Misses:          5%   xxxxxxx  0.00000`))
			} else {
				w.Write([]byte(""))
			}
		default:
			w.Write([]byte(""))
		}
	}))
}

func captureOutput(f func() error) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestRequestHandlerFail(t *testing.T) {
	cases := []struct {
		name        string
		step        string
		expectedLog string
	}{
		{
			name:        "test",
			step:        "counters",
			expectedLog: "error getting squid counters",
		},
		{
			name:        "test2",
			step:        "convert_counters",
			expectedLog: "failed to convert squid counters",
		},
		{
			name:        "test3",
			step:        "service_times",
			expectedLog: "error getting squid service_times",
		},
		{
			name:        "test4",
			step:        "convert_service_times",
			expectedLog: "failed to convert squid service_time",
		},
		{
			name:        "test4",
			step:        "success",
			expectedLog: "successfully converted squid metrics",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts := serverFail(c.step)
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			port, err := strconv.Atoi(u.Port())
			if err != nil {
				t.Fatal(err)
			}
			conf := &Config{
				SquidHost: u.Hostname(),
				SquidPort: port,
			}
			s := NewSquidClient(conf)
			out, _ := captureOutput(func() error {
				logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
				requestHandler(logger, s)
				return nil
			})
			assert.NoError(t, err)
			assert.Contains(t, out, c.expectedLog)
		})
	}
}
