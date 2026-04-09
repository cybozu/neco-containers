package bucketprovisioner_test

import (
	"testing"
	"time"

	"github.com/cybozu-private/pdx-containers/bucket-provisioner-light/bucketprovisioner"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		secretAccessKey string
		sessionToken    string
		s3Endpoint      string
		requestTimeout  time.Duration
		wantErr         bool
		want            *bucketprovisioner.Config
	}{
		{
			name:            "valid with explicit port",
			accessKeyID:     "mykey",
			secretAccessKey: "mysecret",
			sessionToken:    "mytoken",
			s3Endpoint:      "http://localhost:9000",
			requestTimeout:  10 * time.Second,
			wantErr:         false,
			want: &bucketprovisioner.Config{
				AccessKeyID:     "mykey",
				SecretAccessKey: "mysecret",
				SessionToken:    "mytoken",
				Endpoint: bucketprovisioner.Endpoint{
					URL:  "http://localhost:9000",
					Host: "localhost",
					Port: 9000,
				},
				RequestTimeout: 10 * time.Second,
			},
		},
		{
			name:            "valid https without port defaults to 443",
			accessKeyID:     "mykey",
			secretAccessKey: "mysecret",
			sessionToken:    "",
			s3Endpoint:      "https://localhost",
			requestTimeout:  15 * time.Second,
			wantErr:         false,
			want: &bucketprovisioner.Config{
				AccessKeyID:     "mykey",
				SecretAccessKey: "mysecret",
				SessionToken:    "",
				Endpoint: bucketprovisioner.Endpoint{
					URL:  "https://localhost",
					Host: "localhost",
					Port: 443,
				},
				RequestTimeout: 15 * time.Second,
			},
		},
		{
			name:            "empty secretAccessKey returns error",
			accessKeyID:     "mykey",
			secretAccessKey: "",
			sessionToken:    "",
			s3Endpoint:      "http://localhost:9000",
			wantErr:         true,
		},
		{
			name:            "empty endpoint returns error",
			accessKeyID:     "mykey",
			secretAccessKey: "mysecret",
			sessionToken:    "",
			s3Endpoint:      "",
			wantErr:         true,
		},
		{
			name:            "unknown scheme without port returns error",
			accessKeyID:     "mykey",
			secretAccessKey: "mysecret",
			sessionToken:    "",
			s3Endpoint:      "ftp://localhost",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bucketprovisioner.NewConfig(tt.accessKeyID, tt.secretAccessKey, tt.sessionToken, tt.s3Endpoint, tt.requestTimeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.AccessKeyID != tt.want.AccessKeyID {
				t.Errorf("AccessKeyID = %q, want %q", got.AccessKeyID, tt.want.AccessKeyID)
			}
			if got.SecretAccessKey != tt.want.SecretAccessKey {
				t.Errorf("SecretAccessKey = %q, want %q", got.SecretAccessKey, tt.want.SecretAccessKey)
			}
			if got.SessionToken != tt.want.SessionToken {
				t.Errorf("SessionToken = %q, want %q", got.SessionToken, tt.want.SessionToken)
			}
			if got.Endpoint.URL != tt.want.Endpoint.URL {
				t.Errorf("Endpoint.URL = %q, want %q", got.Endpoint.URL, tt.want.Endpoint.URL)
			}
			if got.Endpoint.Host != tt.want.Endpoint.Host {
				t.Errorf("Endpoint.Host = %q, want %q", got.Endpoint.Host, tt.want.Endpoint.Host)
			}
			if got.Endpoint.Port != tt.want.Endpoint.Port {
				t.Errorf("Endpoint.Port = %d, want %d", got.Endpoint.Port, tt.want.Endpoint.Port)
			}
			if got.RequestTimeout != tt.want.RequestTimeout {
				t.Errorf("RequestTimeout = %v, want %v", got.RequestTimeout, tt.want.RequestTimeout)
			}
		})
	}
}
