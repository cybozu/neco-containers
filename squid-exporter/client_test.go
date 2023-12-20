package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSquidClient(t *testing.T) {
	cases := []struct {
		name     string
		config   *Config
		expected *squidClient
	}{
		{
			name: "test1",
			config: &Config{
				SquidHost: "localhost",
				SquidPort: 3128,
			},
			expected: &squidClient{
				client: &http.Client{},
				Host:   "localhost",
				Port:   3128,
			},
		},
		{
			name: "test2",
			config: &Config{
				SquidHost: "127.0.0.1",
				SquidPort: 3128,
			},
			expected: &squidClient{
				client: &http.Client{},
				Host:   "127.0.0.1",
				Port:   3128,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewSquidClient(tc.config)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetGetCounters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.String()))
	}))
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

	res, err := s.GetCounters()
	assert.NoError(t, err)
	resp, err := io.ReadAll(res)
	assert.NoError(t, err)
	assert.Equal(t, "/squid-internal-mgr/counters", string(resp))
}

func TestGetServiceTimes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.URL.String()))
	}))
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

	res, err := s.GetServiceTimes()
	assert.NoError(t, err)
	resp, err := io.ReadAll(res)
	assert.NoError(t, err)
	assert.Equal(t, "/squid-internal-mgr/service_times", string(resp))
}
