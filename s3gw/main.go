package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
)

var bucketHost string
var bucketPort string
var bucketHostPort string
var bucketName string
var bucketRegion string
var awsCredentials aws.Credentials

const metricsNamespace = "s3gw"

var (
	requestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "request_count_total",
			Help:      "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method", "handler"},
	)
	durationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Name:      "request_duration_seconds",
			Help:      "A histogram of latencies for requests.",
			Buckets:   []float64{.25, .5, 1, 2.5, 5, 10, 25, 50, 100},
		},
		[]string{"code", "method", "handler"},
	)
)

func init() {
	prometheus.MustRegister(requestsCounter)
	prometheus.MustRegister(durationHistogram)
}

var flagUsePathStyle bool
var flagListen string
var flagHostsAllow string
var flagHostsDeny string

func init() {
	flag.BoolVar(&flagUsePathStyle, "use-path-style", false, "use path style bucket name")
	flag.StringVar(&flagListen, "listen", ":80", "addr:port to listen to")
	flag.StringVar(&flagHostsAllow, "hosts-allow", "", "subnets allowed to access to this gw, separated by comma")
	flag.StringVar(&flagHostsDeny, "hosts-deny", "", "subnets denied to access to this gw, separated by comma")
}

func main() {
	flag.Parse()
	well.LogConfig{}.Apply()

	// names of envs which must not be empty
	envNames := []string{
		"BUCKET_HOST",
		"BUCKET_NAME",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
	}
	emptyEnvNames := []string{}
	for _, n := range envNames {
		if os.Getenv(n) == "" {
			emptyEnvNames = append(emptyEnvNames, n)
		}
	}
	if len(emptyEnvNames) != 0 {
		log.ErrorExit(fmt.Errorf("some environment variables are empty: %v", emptyEnvNames))
	}

	bucketHost = os.Getenv("BUCKET_HOST")
	bucketName = os.Getenv("BUCKET_NAME")
	bucketPort = os.Getenv("BUCKET_PORT")
	bucketRegion = os.Getenv("BUCKET_REGION")
	awsCredentials.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	awsCredentials.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	if bucketPort == "" {
		bucketHostPort = bucketHost
	} else {
		bucketHostPort = bucketHost + ":" + bucketPort
	}

	allowdeny, err := ParseAllowDeny(flagHostsAllow, flagHostsDeny)
	if err != nil {
		log.ErrorExit(err)
	}

	client = s3.NewFromConfig(aws.Config{
		Region:      bucketRegion,
		Credentials: credentialsProvider{},
	}, func(opts *s3.Options) {
		opts.EndpointResolver = s3.EndpointResolverFromURL("http://" + bucketHostPort + "/")
		opts.UsePathStyle = flagUsePathStyle
	})
	mux := http.NewServeMux()
	server := &well.HTTPServer{
		Server: &http.Server{
			Handler: mux,
			Addr:    flagListen,
		},
	}

	listHandler :=
		promhttp.InstrumentHandlerCounter(requestsCounter.MustCurryWith(prometheus.Labels{"handler": "list"}),
			promhttp.InstrumentHandlerDuration(durationHistogram.MustCurryWith(prometheus.Labels{"handler": "list"}),
				http.HandlerFunc(listHandlerFunc)))
	objectHandler :=
		promhttp.InstrumentHandlerCounter(requestsCounter.MustCurryWith(prometheus.Labels{"handler": "object"}),
			promhttp.InstrumentHandlerDuration(durationHistogram.MustCurryWith(prometheus.Labels{"handler": "object"}),
				http.HandlerFunc(objectHandlerFunc)))
	mux.HandleFunc("/bucket/", func(res http.ResponseWriter, req *http.Request) {
		if !allowdeny.IsAllowedHostPort(req.RemoteAddr) {
			res.WriteHeader(http.StatusForbidden)
			return
		}
		if req.URL.Path == "/bucket/" {
			listHandler(res, req)
		} else {
			objectHandler(res, req)
		}
	})

	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	err = server.ListenAndServe()
	if err != nil {
		log.ErrorExit(err)
	}
	log.Info("starting...", nil)

	well.Stop()
	err = well.Wait()
	if !well.IsSignaled(err) && err != nil {
		log.ErrorExit(err)
	}
}
