package main

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/Hsn723/certspotter-client/api"
	"github.com/Hsn723/ct-monitor/filter"
	"github.com/hashicorp/go-plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	issuerNameLabel = "cert-manager.io/issuer-name"
	issuerKindLabel = "cert-manager.io/issuer-kind"
	targetKind      = "ClusterIssuer"
)

var targetIssuers = []string{
	"clouddns",
	"clouddns-letsencrypt",
}

var certificateRequestGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificaterequests",
}

type inclusterFilter struct {
	client dynamic.Interface
}

func newInclusterFilter() (*inclusterFilter, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return &inclusterFilter{client: client}, nil
}

func (f *inclusterFilter) clusterFingerprints(ctx context.Context) (map[string]struct{}, error) {
	fingerprints := make(map[string]struct{})
	for _, issuer := range targetIssuers {
		list, err := f.client.Resource(certificateRequestGVR).Namespace("").List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", issuerNameLabel, issuer, issuerKindLabel, targetKind),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list CertificateRequests for issuer %s: %w", issuer, err)
		}
		for _, item := range list.Items {
			status, ok := item.Object["status"].(map[string]interface{})
			if !ok {
				continue
			}
			certStr, ok := status["certificate"].(string)
			if !ok || certStr == "" {
				continue
			}
			// dynamic client returns []byte fields as base64-encoded strings
			certPEM, err := base64.StdEncoding.DecodeString(certStr)
			if err != nil {
				continue
			}
			block, _ := pem.Decode(certPEM)
			if block == nil {
				continue
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}
			sum := sha256.Sum256(cert.Raw)
			fingerprints[hex.EncodeToString(sum[:])] = struct{}{}
		}
	}
	return fingerprints, nil
}

func (f *inclusterFilter) Filter(issuances []api.Issuance) ([]api.Issuance, error) {
	fingerprints, err := f.clusterFingerprints(context.Background())
	if err != nil {
		return issuances, err
	}

	filtered := issuances[:0]
	for _, is := range issuances {
		if _, known := fingerprints[is.CertSHA256]; !known {
			filtered = append(filtered, is)
		}
	}
	return filtered, nil
}

func main() {
	f, err := newInclusterFilter()
	if err != nil {
		panic(err)
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: filter.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			filter.PluginKey: &filter.IssuanceFilterPlugin{Impl: f},
		},
	})
}
