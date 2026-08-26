package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
)

var (
	flgSabakanAddress = pflag.String("sabakan-address", "", "address of sabakan's GraphQL API, in the form host:port")

	// flags for monitoring Service/EndpointSlices; the boot-servers/all-servers
	// target is only created when its port list is non-empty.
	flgAllServersPorts      = pflag.StringArray("all-servers-port", []string{}, "port to expose on the all-servers target, in the form port:name (may be repeated); the target is not created if unset")
	flgBootServersPorts     = pflag.StringArray("boot-servers-port", []string{}, "port to expose on the boot-servers target, in the form port:name (may be repeated); the target is not created if unset")
	flgMaxEndpointsPerSlice = pflag.Int("max-endpoints-per-slice", defaultMaxEndpointsPerSlice, "maximum number of endpoints per EndpointSlice")

	// flags for BMC-related ConfigMaps
	flgBMCReverseProxyConfigMap = pflag.Bool("bmc-reverse-proxy-configmap", false, "generate ConfigMap for BMC reverse proxy")
	flgBMCLogCollectorConfigMap = pflag.Bool("bmc-log-collector-configmap", false, "generate ConfigMap for BMC log collector")

	// flag for dry-run
	flgDryRun = pflag.Bool("dry-run", false, "print resources to stdout instead of updating them in Kubernetes")
)

func main() {
	pflag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := subMain(ctx); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func subMain(ctx context.Context) error {
	allServersPorts, err := parseNamedPorts(*flgAllServersPorts)
	if err != nil {
		return fmt.Errorf("all-servers-port: %w", err)
	}
	bootServersPorts, err := parseNamedPorts(*flgBootServersPorts)
	if err != nil {
		return fmt.Errorf("boot-servers-port: %w", err)
	}
	if err := validateMaxEndpointsPerSlice(*flgMaxEndpointsPerSlice); err != nil {
		return err
	}

	sabakanHost, sabakanPort, err := net.SplitHostPort(*flgSabakanAddress)
	if err != nil || sabakanHost == "" || sabakanPort == "" {
		return fmt.Errorf("invalid sabakan-address: %q (expected host:port)", *flgSabakanAddress)
	}
	sabakanCli := newSabakanClient(*flgSabakanAddress)

	machines, err := sabakanCli.getMachines(ctx)
	if err != nil {
		return err
	}

	k8sCli, err := newKubernetesClient(*flgDryRun)
	if err != nil {
		return err
	}

	if len(allServersPorts) > 0 {
		ips := allServerIPs(machines)
		if err := k8sCli.updateTargetEndpoints(ctx, targetAllServersName, ips, *flgMaxEndpointsPerSlice, allServersPorts); err != nil {
			return err
		}
	}

	if len(bootServersPorts) > 0 {
		ips := bootServerIPs(machines)
		if err := k8sCli.updateTargetEndpoints(ctx, targetBootServersName, ips, *flgMaxEndpointsPerSlice, bootServersPorts); err != nil {
			return err
		}
	}

	if *flgBMCReverseProxyConfigMap {
		data := bmcReverseProxyConfigMapData(machines)
		if err := k8sCli.applyConfigMap(ctx, bmcReverseProxyConfigMapName, data); err != nil {
			return err
		}
	}

	if *flgBMCLogCollectorConfigMap {
		data, err := bmcLogCollectorConfigMapData(machines)
		if err != nil {
			return err
		}
		if err := k8sCli.applyConfigMap(ctx, bmcLogCollectorConfigMapName, data); err != nil {
			return err
		}
	}

	return nil
}
