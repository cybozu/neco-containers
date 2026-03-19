package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/Hsn723/certspotter-client/api"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func generateCert(t *testing.T) (*x509.Certificate, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return mustParseCert(t, certDER), base64.StdEncoding.EncodeToString(pemBytes)
}

func mustParseCert(t *testing.T, der []byte) *x509.Certificate {
	t.Helper()
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func certFingerprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sum[:])
}

func makeCertificateRequest(namespace, name, issuerName string, certBase64 string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "CertificateRequest",
	})
	obj.SetNamespace(namespace)
	obj.SetName(name)
	obj.SetLabels(map[string]string{
		issuerNameLabel: issuerName,
		issuerKindLabel: targetKind,
	})
	if certBase64 != "" {
		_ = unstructured.SetNestedField(obj.Object, certBase64, "status", "certificate")
	}
	return obj
}

func newTestFilter(t *testing.T, objs ...runtime.Object) *inclusterFilter {
	t.Helper()
	scheme := runtime.NewScheme()
	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			certificateRequestGVR: "CertificateRequestList",
		},
		objs...,
	)
	return &inclusterFilter{client: client}
}

func TestClusterFingerprints_Empty(t *testing.T) {
	t.Parallel()
	f := newTestFilter(t)
	fps, err := f.clusterFingerprints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fps) != 0 {
		t.Errorf("expected empty fingerprints, got %d", len(fps))
	}
}

func TestClusterFingerprints_ValidCert(t *testing.T) {
	t.Parallel()
	cert, certBase64 := generateCert(t)
	expected := certFingerprint(cert)

	f := newTestFilter(t,
		makeCertificateRequest("default", "cr-1", "clouddns", certBase64),
	)
	fps, err := f.clusterFingerprints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := fps[expected]; !ok {
		t.Errorf("expected fingerprint %s not found", expected)
	}
}

func TestClusterFingerprints_MultipleIssuers(t *testing.T) {
	t.Parallel()
	cert1, certBase64_1 := generateCert(t)
	cert2, certBase64_2 := generateCert(t)

	f := newTestFilter(t,
		makeCertificateRequest("ns1", "cr-1", "clouddns", certBase64_1),
		makeCertificateRequest("ns2", "cr-2", "clouddns-letsencrypt", certBase64_2),
	)
	fps, err := f.clusterFingerprints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := fps[certFingerprint(cert1)]; !ok {
		t.Error("clouddns cert fingerprint not found")
	}
	if _, ok := fps[certFingerprint(cert2)]; !ok {
		t.Error("clouddns-letsencrypt cert fingerprint not found")
	}
}

func TestClusterFingerprints_NoCertificate(t *testing.T) {
	t.Parallel()
	f := newTestFilter(t,
		makeCertificateRequest("default", "cr-1", "clouddns", ""),
	)
	fps, err := f.clusterFingerprints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fps) != 0 {
		t.Errorf("expected empty fingerprints for pending CR, got %d", len(fps))
	}
}

func TestClusterFingerprints_InvalidBase64(t *testing.T) {
	t.Parallel()
	cr := makeCertificateRequest("default", "cr-1", "clouddns", "not-valid-base64!!!")
	f := newTestFilter(t, cr)
	fps, err := f.clusterFingerprints(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fps) != 0 {
		t.Errorf("expected invalid base64 to be skipped, got %d fingerprints", len(fps))
	}
}

func TestFilter_RemovesKnownIssuances(t *testing.T) {
	t.Parallel()
	cert, certBase64 := generateCert(t)
	fp := certFingerprint(cert)

	f := newTestFilter(t,
		makeCertificateRequest("default", "cr-1", "clouddns", certBase64),
	)
	issuances := []api.Issuance{
		{CertSHA256: fp},
		{CertSHA256: "unknown-fingerprint"},
	}
	result, err := f.Filter(issuances)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 issuance, got %d", len(result))
	}
	if result[0].CertSHA256 != "unknown-fingerprint" {
		t.Errorf("expected unknown-fingerprint to remain, got %s", result[0].CertSHA256)
	}
}

func TestFilter_PassesAllWhenNoClusterCerts(t *testing.T) {
	t.Parallel()
	f := newTestFilter(t)
	issuances := []api.Issuance{
		{CertSHA256: "fp-1"},
		{CertSHA256: "fp-2"},
	}
	result, err := f.Filter(issuances)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 issuances, got %d", len(result))
	}
}

func TestFilter_RemovesAllWhenAllKnown(t *testing.T) {
	t.Parallel()
	cert, certBase64 := generateCert(t)

	f := newTestFilter(t,
		makeCertificateRequest("default", "cr-1", "clouddns", certBase64),
	)
	issuances := []api.Issuance{
		{CertSHA256: certFingerprint(cert)},
	}
	result, err := f.Filter(issuances)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 issuances, got %d", len(result))
	}
}
