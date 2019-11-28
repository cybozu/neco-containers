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
)

const graphQLQuery = `
query search() {
  searchMachines(having: null, notHaving: null) {
    spec {
      ipv4
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
		IPv4 []string `json:"ipv4"`
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

type machineInfo struct {
	ip net.IP
}

func (c client) getMachinesFromSabakan(bootservers []net.IP) ([]machineInfo, error) {
	if len(bootservers) == 0 {
		return nil, errors.New("no bootservers")
	}

	var machines []machineInfo
	var err error
	for _, boot := range bootservers {
		machines, err = func() ([]machineInfo, error) {
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

			for _, machine := range result.Data.Machines {
				if machine.Spec.IPv4 == nil {
					continue
				}
				if len(machine.Spec.IPv4) == 0 {
					continue
				}
				machines = append(machines, machineInfo{
					ip: net.ParseIP(machine.Spec.IPv4[0])})
			}
			return machines, nil
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

		machineIPs := make([]net.IP, len(machines))
		for idx, machine := range machines {
			machineIPs[idx] = machine.ip
		}

		// create node-exporter endpoints on all servers
		err = client.updateTargetEndpoints(machineIPs, targetEndpointsName, nodeExporterPortName, *flgNodeExporterPort)
		if err != nil {
			log.ErrorExit(err)
		}
	}

	if *flgBMCConfigMap {
	}
}
