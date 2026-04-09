package bucketprovisioner

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config holds the validated configuration for the S3 provisioner.
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        Endpoint
	RequestTimeout  time.Duration
}

type Endpoint struct {
	URL  string
	Host string
	Port int
}

// NewConfig validates the given parameters and returns a Config.
func NewConfig(accessKeyID, secretAccessKey, sessionToken, s3Endpoint string, requestTimeout time.Duration) (Config, error) {
	if accessKeyID == "" || secretAccessKey == "" {
		return Config{}, errors.New("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set")
	}

	e := strings.TrimSpace(s3Endpoint)
	if e == "" {
		return Config{}, errors.New("S3_ENDPOINT must be set")
	}

	endpointURL, err := url.Parse(e)
	if err != nil {
		return Config{}, fmt.Errorf("parse S3_ENDPOINT: %w", err)
	}

	port, err := strconv.Atoi(endpointURL.Port())
	if err != nil {
		switch endpointURL.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			return Config{}, fmt.Errorf("invalid port or scheme in S3_ENDPOINT: %w", err)
		}
	}

	return Config{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		Endpoint: Endpoint{
			URL:  endpointURL.String(),
			Host: endpointURL.Hostname(),
			Port: port,
		},
		RequestTimeout: requestTimeout,
	}, nil
}
