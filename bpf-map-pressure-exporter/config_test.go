package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		name        string
		path        string
		expect      *Config
		expectError bool
	}{
		{
			name: "success",
			path: "testdata/config.yaml",
			expect: &Config{
				MapNames:      []string{"hoge", "fuga", "piyo"},
				FetchInterval: 1 * time.Minute,
			},
		},
		{
			name: "default fetch interval",
			path: "testdata/no-fetch-interval.yaml",
			expect: &Config{
				MapNames:      []string{"hoge", "fuga", "piyo"},
				FetchInterval: defaultFetchInterval,
			},
		},
		{
			name:        "file not found",
			path:        "testdata/notfound.yaml",
			expectError: true,
		},
		{
			name:        "invalid yaml",
			path:        "testdata/invalid.yaml",
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := loadConfig(tc.path)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expect.MapNames, cfg.MapNames)
		})
	}
}
