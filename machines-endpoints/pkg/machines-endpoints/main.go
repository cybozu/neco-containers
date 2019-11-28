package main

import (
	"bytes"
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
	serfclient "github.com/hashicorp/serf/client"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
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
	bmcProxyConfigMapName = "bmc-proxy"
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
  }
}
`

var (
	flgMonitoringEndpoints = pflag.Bool("monitoring-endpoints", false, "generate Endpoints for monitoring")
	flgBMCConfigMap        = pflag.Bool("bmc-configmap", false, "generate ConfigMap for BMC reverse proxy")
	flgNodeExporterPort    = pflag.Int32("node-exporter-port", defaultNodeExporterPort, "node-exporter port")
	flgEtcdMetricsPort     = pflag.Int32("etcd-metrics-port", defaultEtcdMetricsPort, "etcd metrics port")
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

func (c client) updateTargetEndpoints(targetIPs []net.IP, target, portName string, port int32) error {
	ns, _, err := c.kubeConfig.Namespace()
	if err != nil {
		return err
	}

	services := c.k8s.CoreV1().Services(ns)
	_, err = services.Get(target, metav1.GetOptions{})
	switch {
	case err == nil:
	case k8serrors.IsNotFound(err):
		_, err = services.Create(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: target,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{Port: port, TargetPort: intstr.FromInt(int(port)), Name: portName},
				},
				ClusterIP: "None",
			},
		})
		if err != nil {
			return err
		}
	default:
		return err
	}

	subset := corev1.EndpointSubset{
		Addresses: make([]corev1.EndpointAddress, len(targetIPs)),
		Ports: []corev1.EndpointPort{
			{Port: port, Name: portName},
		},
	}
	for i, ip := range targetIPs {
		subset.Addresses[i].IP = ip.String()
	}

	_, err = c.k8s.CoreV1().Endpoints(ns).Update(&corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: target,
		},
		Subsets: []corev1.EndpointSubset{subset},
	})
	return err
}

func (c client) updateBMCProxyConfigMap(machines []Machine) error {
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
	_, err = cms.Get(bmcProxyConfigMapName, metav1.GetOptions{})
	switch {
	case err == nil:
		_, err := cms.Update(&configMap)
		return err
	case k8serrors.IsNotFound(err):
		_, err := cms.Create(&configMap)
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
	for _, s := range serfMembers {
		members = append(members, member{name: s.Name, addr: s.Addr, tags: s.Tags})
	}
	return members, nil
}

func getBootServers(members *[]member) []net.IP {
	var bootservers []net.IP
	for _, member := range *members {
		if member.tags["boot-server"] == "true" {
			bootservers = append(bootservers, member.addr)
		}
	}
	return bootservers
}

func localHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
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

	bootservers := getBootServers(&members)
	machines, err := client.getMachinesFromSabakan(bootservers)
	if err != nil {
		log.ErrorExit(err)
	}

	if *flgMonitoringEndpoints {
		// create etcd metrics endpoints on boot servers
		err = client.updateTargetEndpoints(bootservers, targetEtcdMetricsEndpointsName, etcdMetricsPortName, *flgEtcdMetricsPort)
		if err != nil {
			log.ErrorExit(err)
		}

		machineIPs := make([]net.IP, 0, len(machines))
		for _, machine := range machines {
			if machine.Spec.IPv4 == nil {
				continue
			}
			if len(machine.Spec.IPv4) == 0 {
				continue
			}
			machineIPs = append(machineIPs, net.ParseIP(machine.Spec.IPv4[0]))
		}

		// create node-exporter endpoints on all servers
		err = client.updateTargetEndpoints(machineIPs, targetEndpointsName, nodeExporterPortName, *flgNodeExporterPort)
		if err != nil {
			log.ErrorExit(err)
		}
	}

	if *flgBMCConfigMap {
		// create bmc-proxy configmap on all servers
		err = client.updateBMCProxyConfigMap(machines)
		if err != nil {
			log.ErrorExit(err)
		}
	}
}
