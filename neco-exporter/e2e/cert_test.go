package e2e

import (
	"fmt"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func testCertCollector() {
	secretName := "cert-dummy"

	It("should add dummy certificate", func() {
		keyPath := fmt.Sprintf("/tmp/%s.key", secretName)
		crtPath := fmt.Sprintf("/tmp/%s.crt", secretName)

		Eventually(func(g Gomega) {
			secrets := kubectlGetSafe[corev1.SecretList](g, "secret", "-n=default")
			if slices.ContainsFunc(secrets.Items, func(s corev1.Secret) bool { return s.Name == secretName }) {
				return
			}
			pilotSafe(
				g, nil, "openssl", "req", "-x509", "-newkey=ed25519", "-sha256", "-nodes", "-days=123", "-subj=/CN=localhost", "-keyout="+keyPath, "-out="+crtPath,
			)
			pilotSafe(
				g, nil, "/tmp/kubectl", "create", "secret", "tls", "-n=default", secretName, "--cert=/tmp/cert-dummy.key", "--key="+keyPath, "--cert="+crtPath,
			)
		}).Should(Succeed())
	})

	It("should export cert expiration date", func() {
		Eventually(func(g Gomega) {
			s := kubectlGetSafe[corev1.Secret](g, "secret", "-n=default", secretName)
			crt := s.Data["tls.crt"]

			// e.g. notAfter=May 29 02:24:14 2026 GMT
			line := string(pilotSafe(g, crt, "openssl", "x509", "-noout", "-enddate"))
			parts := strings.Split(line, "=")
			g.Expect(parts).To(HaveLen(2))

			const layout = "Jan _2 15:04:05 2006 MST"
			expiration, err := time.Parse(layout, strings.TrimSpace(parts[1]))
			g.Expect(err).NotTo(HaveOccurred())

			// ref. https://github.com/VictoriaMetrics/metrics/blob/v1.40.1/floatcounter.go#L63
			expected := fmt.Sprintf("%g", float64(expiration.UnixNano()))

			found := false
			for li := range strings.Lines(string(scrapeClusterLeader(g))) {
				if strings.Contains(li, "neco_cluster_cert_expiration_timestamp_seconds") &&
					strings.Contains(li, secretName) {
					found = true

					fields := strings.Fields(li)
					g.Expect(fields).To(HaveLen(2))
					g.Expect(fields[1]).To(Equal(expected))
				}
			}
			g.Expect(found).To(BeTrue())
		}).Should(Succeed())
	})

	It("should remove metrics for deleted certificate", func() {
		Eventually(func(g Gomega) {
			secrets := kubectlGetSafe[corev1.SecretList](g, "secret", "-n=default")
			if !slices.ContainsFunc(secrets.Items, func(s corev1.Secret) bool { return s.Name == secretName }) {
				return
			}
			kubectlSafe(g, nil, "delete", "secret", secretName)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			found := false
			for li := range strings.Lines(string(scrapeClusterLeader(g))) {
				if strings.Contains(li, "neco_cluster_cert_expiration_timestamp_seconds") &&
					strings.Contains(li, secretName) {
					found = true
				}
			}
			g.Expect(found).To(BeFalse())
		}).Should(Succeed())
	})
}
