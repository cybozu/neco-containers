package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	sabakan "github.com/cybozu-go/sabakan/v3"
	serfclient "github.com/hashicorp/serf/client"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// node-exporter
	targetEndpointsName     = "prometheus-node-targets"
	nodeExporterPortName    = "http-node-exporter"
	defaultNodeExporterPort = 9100

	// etcd metrics
	targetEtcdMetricsEndpointsName = "bootserver-etcd-metrics"
	etcdMetricsPortName            = "http-etcd-metrics"
	defaultEtcdMetricsPort         = 2381

	// BMC Proxy ConfigMap
	bmcProxyConfigMapName = "bmc-reverse-proxy"

	// BMC Log Collector ConfigMap
	bmcLogCollectorConfigMapName = "bmc-log-collector"	
)

const graphQLQuery = `
query search() {
  searchMachines(having: null, notHaving: null) {
    spec {
      ipv4
      serial
      rack
      indexInRack
      role
      bmc {
        ipv4
      }
    }
    status {
      state
    }
  }
}
`

var (
	flgMonitoringEndpoints      = pflag.Bool("monitoring-endpoints", false, "generate Endpoints for monitoring")
	flgBMCReverseProxyConfigMap = pflag.Bool("bmc-configmap", false, "generate ConfigMap for BMC reverse proxy")
	flgBMCLogCollectorConfigMap = pflag.Bool("log-collector", false, "generate ConfigMap for BMC log collector")
	flgNodeExporterPort         = pflag.Int32("node-exporter-port", defaultNodeExporterPort, "node-exporter port")
	flgEtcdMetricsPort          = pflag.Int32("etcd-metrics-port", defaultEtcdMetricsPort, "etcd metrics port")
)

// Machine represents a machine registered with sabakan.
type Machine struct {
	Spec struct {
		IPv4        []string `json:"ipv4"`
		Serial      string   `json:"serial"`
		Rack        int      `json:"rack"`
		IndexInRack int      `json:"indexInRack"`
		Role        string   `json:"role"`
		BMC         struct {
			IPv4 string `json:"ipv4"`
		}
	}
	Status struct {
		State string `json:"state"`
	}
}

type member struct {
	name string
	addr net.IP
	tags map[string]string
}

type client struct {
	http       *http.Client
	k8s        *kubernetes.Clientset
	kubeConfig clientcmd.ClientConfig
	serf       *serfclient.RPCClient
}

func (c client) getMachinesFromSabakan(bootservers []net.IP) ([]Machine, error) {
	if len(bootservers) == 0 {
		return nil, errors.New("no bootservers")
	}

	var machines []Machine
	var err error
	for _, boot := range bootservers {
		machines, err = func() ([]Machine, error) {
			addr := net.JoinHostPort(boot.String(), "10080")
			queryURL, err := url.Parse("http://" + addr)
			if err != nil {
				return nil, err
			}

			queryURL.Path = path.Join(queryURL.Path, "/graphql")
			body := struct {
				Query string `json:"query"`
			}{
				graphQLQuery,
			}
			data, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}

			req, err := http.NewRequest(http.MethodPost, queryURL.String(), bytes.NewReader(data))
			if err != nil {
				return nil, err
			}
			// gqlgen 0.9+ requires application/json content-type header.
			req.Header.Set("Content-Type", "application/json")
			resp, err := c.http.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var result struct {
				Data struct {
					Machines []Machine `json:"searchMachines"`
				} `json:"data"`
				Errors []struct {
					Message string `json:"message"`
				} `json:"errors"`
			}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				return nil, err
			}
			if len(result.Errors) > 0 {
				return nil, fmt.Errorf("sabakan returned error: %v", result.Errors)
			}

			return result.Data.Machines, nil
		}()
		if err == nil {
			return machines, nil
		}
		log.Error("failed to get machines from sabakan", map[string]interface{}{
			"bootserver": boot.String(),
			log.FnError:  err,
		})
	}
	return nil, err
}

func (c client) updateTargetEndpoints(ctx context.Context, targetIPs []net.IP, target, portName string, port int32) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	// Create Service if not exists
	services := c.k8s.CoreV1().Services(ns)
	_, err = services.Get(ctx, target, metav1.GetOptions{})
	switch {
	case err == nil:
	case k8serrors.IsNotFound(err):
		_, err = services.Create(ctx, &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: target,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{Port: port, TargetPort: intstr.FromInt(int(port)), Name: portName},
				},
				ClusterIP: "None",
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	default:
		return err
	}

	// Create or update Endpoints
	subset := corev1.EndpointSubset{
		Addresses: make([]corev1.EndpointAddress, len(targetIPs)),
		Ports: []corev1.EndpointPort{
			{Port: port, Name: portName},
		},
	}
	for i, ip := range targetIPs {
		subset.Addresses[i].IP = ip.String()
	}
	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: target,
			Labels: map[string]string{
				"endpointslice.kubernetes.io/skip-mirror": "true",
			},
		},
		Subsets: []corev1.EndpointSubset{subset},
	}
	endpointInterface := c.k8s.CoreV1().Endpoints(ns)
	_, err = endpointInterface.Update(ctx, ep, metav1.UpdateOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = endpointInterface.Create(ctx, ep, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	// Create or update EndpointSlice
	endpoints := make([]discoveryv1.Endpoint, len(targetIPs))
	for i, ip := range targetIPs {
		endpoints[i] = discoveryv1.Endpoint{Addresses: []string{ip.String()}}
	}
	endpointSlice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: target,
			Labels: map[string]string{
				"endpointslice.kubernetes.io/managed-by": "machines-endpoints.cybozu.com",
				"kubernetes.io/service-name":             target,
			},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints:   endpoints,
		Ports: []discoveryv1.EndpointPort{
			{
				Name: &portName,
				Port: &port,
			},
		},
	}
	endpointSliceInterface := c.k8s.DiscoveryV1().EndpointSlices(ns)
	_, err = endpointSliceInterface.Update(ctx, endpointSlice, metav1.UpdateOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = endpointSliceInterface.Create(ctx, endpointSlice, metav1.CreateOptions{})
	}
	return err
}

func (c client) updateBMCProxyConfigMap(ctx context.Context, machines []Machine) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	addresses := make(map[string]string)
	for _, machine := range machines {
		if machine.Spec.BMC.IPv4 == "" {
			continue
		}

		var hostname string
		if machine.Spec.Role == "boot" {
			// Though full hostname is like "stage0-boot-0",
			// the part of "stage0-" is insignificant in a cluster while it is hard to get.
			// So use "boot-0" for resolving.
			hostname = fmt.Sprintf("boot-%d", machine.Spec.Rack)
		} else {
			hostname = fmt.Sprintf("rack%d-%s%d", machine.Spec.Rack, machine.Spec.Role, machine.Spec.IndexInRack)
		}
		addresses[hostname] = machine.Spec.BMC.IPv4

		// "a.b.c.d" does not match the wildcard in "*.bmc.<cluster>.<base>".  "a-b-c-d" does match.
		addresses[strings.ReplaceAll(machine.Spec.IPv4[0], ".", "-")] = machine.Spec.BMC.IPv4

		addresses[machine.Spec.Serial] = machine.Spec.BMC.IPv4
	}

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: bmcProxyConfigMapName,
		},
		Data: addresses,
	}

	cms := c.k8s.CoreV1().ConfigMaps(ns)
	_, err = cms.Get(ctx, bmcProxyConfigMapName, metav1.GetOptions{})
	switch {
	case err == nil:
		_, err := cms.Update(ctx, &configMap, metav1.UpdateOptions{})
		return err
	case k8serrors.IsNotFound(err):
		_, err := cms.Create(ctx, &configMap, metav1.CreateOptions{})
		return err
	}
	return err
}

func (c client) updateBMCLogCollectorConfigMap(ctx context.Context, machines []Machine) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	type Machine struct {
		Serial string `json:"serial"`
		BmcIP  string `json:"bmc_ip"`
		NodeIP string `json:"ipv4"`
	}

	var ml []Machine
	var m Machine
	for _, machine := range machines {
		if machine.Spec.BMC.IPv4 == "" {
			continue
		}
		m.Serial = machine.Spec.Serial
		m.BmcIP  = machine.Spec.BMC.IPv4
		m.NodeIP = machine.Spec.IPv4[0]
		ml = append(ml, m)
	}

	byteJSON, err := json.Marshal(ml)
	if err != nil {
		return err
	}
	addresses := make(map[string]string)
	addresses["serverlist.json"] = string(byteJSON)

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: bmcLogCollectorConfigMapName,
		},
		Data: addresses,
	}

	cms := c.k8s.CoreV1().ConfigMaps(ns)
	_, err = cms.Get(ctx, bmcLogCollectorConfigMapName, metav1.GetOptions{})
	switch {
	case err == nil:
		_, err := cms.Update(ctx, &configMap, metav1.UpdateOptions{})
		return err
	case k8serrors.IsNotFound(err):
		_, err := cms.Create(ctx, &configMap, metav1.CreateOptions{})
		return err
	}
	return err
}

func (c client) GetMembers() ([]member, error) {
	serfMembers, err := c.serf.Members()
	if err != nil {
		return nil, err
	}

	members := make([]member, len(serfMembers))
	for i, s := range serfMembers {
		members[i] = member{name: s.Name, addr: s.Addr, tags: s.Tags}
	}
	return members, nil
}

func getBootServers(members []member) []net.IP {
	var bootservers []net.IP
	for _, member := range members {
		if member.tags["boot-server"] == "true" {
			bootservers = append(bootservers, member.addr)
		}
	}
	return bootservers
}

func localHTTPClient() *http.Client {
	// Most of the following values are copied from http.DefaultTransport to workaround a proxy issue.
	// See: https://github.com/golang/go/issues/25793
	transport := &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Minute,
	}
}

func main() {
	pflag.Parse()

	serfc, err := serfclient.NewRPCClient("127.0.0.1:7373")
	if err != nil {
		log.ErrorExit(err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.ErrorExit(err)
	}

	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.ErrorExit(err)
	}
	client := client{http: localHTTPClient(), k8s: k8sClientSet, serf: serfc, kubeConfig: kubeConfig}

	members, err := client.GetMembers()
	if err != nil {
		log.ErrorExit(err)
	}

	bootservers := getBootServers(members)
	machines, err := client.getMachinesFromSabakan(bootservers)
	if err != nil {
		log.ErrorExit(err)
	}

	ctx := context.Background()

	if *flgMonitoringEndpoints {
		machineIPs := make([]net.IP, 0, len(machines))
		currentBootservers := make([]net.IP, 0, len(bootservers))
		for _, machine := range machines {
			if machine.Status.State == sabakan.StateRetired.GQLEnum() {
				continue
			}
			if len(machine.Spec.IPv4) == 0 {
				continue
			}
			machineIP := net.ParseIP(machine.Spec.IPv4[0])
			machineIPs = append(machineIPs, machineIP)
			if machine.Spec.Role == "boot" {
				currentBootservers = append(currentBootservers, machineIP)
			}
		}

		// create etcd metrics endpoints on boot servers
		err = client.updateTargetEndpoints(ctx, currentBootservers, targetEtcdMetricsEndpointsName, etcdMetricsPortName, *flgEtcdMetricsPort)
		if err != nil {
			log.ErrorExit(err)
		}

		// create node-exporter endpoints on all servers
		err = client.updateTargetEndpoints(ctx, machineIPs, targetEndpointsName, nodeExporterPortName, *flgNodeExporterPort)
		if err != nil {
			log.ErrorExit(err)
		}
	}

	if *flgBMCReverseProxyConfigMap {
		// create bmc-proxy configmap on all servers
		err = client.updateBMCProxyConfigMap(ctx, machines)
		if err != nil {
			log.ErrorExit(err)
		}
	}

	if *flgBMCLogCollectorConfigMap {
		// create BMC & Server list configmap on all servers
		err = client.updateBMCLogCollectorConfigMap(ctx, machines)
		if err != nil {
			log.ErrorExit(err)
		}
	}
}
