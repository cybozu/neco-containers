package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	updateCacheInterval = 60 * time.Second
	certFileName        = "/etc/bmc-reverse-proxy/tls.crt"
	keyFileName         = "/etc/bmc-reverse-proxy/tls.key"
	// BMC Proxy ConfigMap
	bmcProxyConfigMapName = "bmc-reverse-proxy"
)

var (
	// *tls.Certificate
	certificateCache atomic.Value
	// map[string]string
	resolveMapCache atomic.Value

	k8s        *kubernetes.Clientset
	kubeConfig clientcmd.ClientConfig
)

func getHostname(fqdn string) string {
	hostname := strings.Split(fqdn, ".")[0]
	// If full hostname is like "stage0-boot-0", trim the part of "stage0-" as machines-endpoints does.
	re := regexp.MustCompile(`-(boot-\d+)$`)
	submatches := re.FindStringSubmatch(hostname)
	if len(submatches) >= 2 {
		hostname = submatches[1]
	}
	return hostname
}

func directorToInner(request *http.Request, inner uint16, resolveMap map[string]string) {
	hostname := getHostname(request.Host)
	address, ok := resolveMap[hostname]
	if !ok {
		address, ok = resolveMap[strings.ToUpper(hostname)]
	}
	if !ok {
		log.Error("failed to resolve hostname", map[string]interface{}{
			"hostname": hostname,
		})
		// Director cannot return an error. Set an invalid address to fail.
		address = "0.0.0.0"
	}
	request.URL.Host = fmt.Sprintf("%s:%d", address, inner)
	request.URL.Scheme = "https"
}

func makeTunnel(inner uint16, external uint16) error {
	director := func(request *http.Request) {
		var resolveMap map[string]string
		c := resolveMapCache.Load()
		if c != nil {
			resolveMap = c.(map[string]string)
		} else {
			log.Error("resolve map is not set", nil)
			// Continue with nil resolveMap
		}
		directorToInner(request, inner, resolveMap)
	}

	transport := &http.Transport{
		// copied from http.DefaultTransport with adding TLSClientConfig
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	rp := &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}

	// do similar to well.HTTPServer.ListenAndServeTLS() with replacing Certificates with GetCertificate
	server := &well.HTTPServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", external),
			Handler: rp,
			TLSConfig: &tls.Config{
				NextProtos:               []string{"h2", "http/1.1"},
				GetCertificate:           getCertificate,
				PreferServerCipherSuites: true,
				ClientSessionCache:       tls.NewLRUClientSessionCache(0),
			},
		},
	}

	ln, err := net.Listen("tcp", server.Server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %v", server.Server.Addr, err)
	}

	tlsListener := tls.NewListener(ln, server.Server.TLSConfig)
	return server.Serve(tlsListener)
}

func updateCertificate() error {
	c, err := tls.LoadX509KeyPair(certFileName, keyFileName)
	if err != nil {
		return fmt.Errorf("failed to load x509 key pair from (%s, %s): %v", certFileName, keyFileName, err)
	}
	certificateCache.Store(&c)
	return nil
}

func updateResolveMap(ctx context.Context) error {
	ns, _, err := kubeConfig.Namespace()
	if err != nil {
		return fmt.Errorf("failed to get namespace: %v", err)
	}

	cm, err := k8s.CoreV1().ConfigMaps(ns).Get(ctx, bmcProxyConfigMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap %q: %v", bmcProxyConfigMapName, err)
	}
	resolveMapCache.Store(cm.Data)

	return nil
}

func getCertificate(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
	c := certificateCache.Load()
	if c == nil {
		return nil, errors.New("certificate is not loaded")
	}
	return c.(*tls.Certificate), nil
}

func main() {
	well.LogConfig{}.Apply()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.ErrorExit(fmt.Errorf("failed to get k8s client config: %v", err))
	}

	k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.ErrorExit(fmt.Errorf("failed to get k8s client: %v", err))
	}

	well.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(updateCacheInterval)
		defer ticker.Stop()

		for {
			err := updateCertificate()
			if err != nil {
				return err
			}

			err = updateResolveMap(ctx)
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
	})

	well.Go(func(ctx context.Context) error {
		return makeTunnel(443, 8443)
	})

	well.Go(func(ctx context.Context) error {
		return makeTunnel(5900, 5900)
	})

	well.Stop()
	err = well.Wait()
	if !well.IsSignaled(err) && err != nil {
		log.ErrorExit(err)
	}
}
