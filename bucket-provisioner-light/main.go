package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	libprovisioner "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/cybozu-private/pdx-containers/bucket-provisioner-light/bucketprovisioner"
)

const (
	defaultProbeAddr   = ":8081"
	defaultProvisioner = "bucket-provisioner-light"
	defaultS3Timeout   = 15 * time.Second
)

func main() {
	var accessKeyID string
	var probeAddr string
	var provisionerName string
	var requestTimeout time.Duration
	var secretAccessKey string
	var sessionToken string
	var s3Endpoint string
	var watchNamespace string

	flag.StringVar(&probeAddr, "health-probe-bind-address", defaultProbeAddr, "The address the healthz endpoint binds to.")
	flag.StringVar(&provisionerName, "provisioner-name", defaultProvisioner, "Provisioner name used by ObjectBucketClaims.")
	flag.StringVar(&accessKeyID, "aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID.")
	flag.StringVar(&secretAccessKey, "aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key.")
	flag.StringVar(&sessionToken, "aws-session-token", os.Getenv("AWS_SESSION_TOKEN"), "AWS session token.")
	flag.StringVar(&s3Endpoint, "s3-endpoint", os.Getenv("S3_ENDPOINT"), "S3 endpoint URL.")
	flag.DurationVar(&requestTimeout, "s3-request-timeout", defaultS3Timeout, "Timeout for S3 requests.")
	flag.StringVar(&watchNamespace, "watch-namespace", os.Getenv("WATCH_NAMESPACE"), "Namespace to watch for ObjectBucketClaims.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	cfg, err := bucketprovisioner.NewConfig(accessKeyID, secretAccessKey, sessionToken, s3Endpoint, requestTimeout)
	if err != nil {
		panic(err)
	}

	prov, err := bucketprovisioner.New(context.Background(), cfg, ctrl.Log.WithName("bucketprovisioner"))
	if err != nil {
		panic(err)
	}

	libProv, err := libprovisioner.NewProvisioner(ctrl.GetConfigOrDie(), provisionerName, prov, watchNamespace)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(probeAddr, mux) //nolint:errcheck

	signalCtx := ctrl.SetupSignalHandler()
	go func() {
		<-signalCtx.Done()
		prov.Stop()
	}()

	if err := libProv.RunWithContext(signalCtx); err != nil {
		panic(err)
	}
}
