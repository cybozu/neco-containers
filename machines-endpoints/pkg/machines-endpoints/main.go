package main

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	serfclient "github.com/hashicorp/serf/client"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	monitoringNamespace  = "monitoring"
	targetEndpointsName  = "prometheus-node-targets"
	nodeExporterPortName = "http-node-exporter"
	nodeExporterPort     = 9100
	ckeEtcdPortName      = "http-cke-etcd"
	ckeEtcdPort          = 2381
)

type member struct {
	name string
	addr net.IP
	tags map[string]string
}

type client struct {
	http *http.Client
	k8s  *kubernetes.Clientset
	serf *serfclient.RPCClient
}

func (c client) getMachinesFromSabakan(bootservers []net.IP) (*[]sabakan.Machine, error) {
	if len(bootservers) == 0 {
		return nil, errors.New("no bootservers")
	}

	var machines *[]sabakan.Machine
	var err error
	for _, boot := range bootservers {
		machines, err = func() (*[]sabakan.Machine, error) {
			machines := new([]sabakan.Machine)
			addr := net.JoinHostPort(boot.String(), "10080")
			base, err := url.Parse("http://" + addr)
			if err != nil {
				log.Error("invalid sabakan server address", map[string]interface{}{
					"sabakanaddr": addr,
					log.FnError:   err,
				})
				return nil, err
			}

			base.Path = path.Join(base.Path, "/api/v1/machines")
			req, err := http.NewRequest(http.MethodGet, base.String(), nil)
			if err != nil {
				return nil, err
			}
			resp, err := c.http.Do(req)
			if err != nil {
				log.Error("failed to get machines from sabakan server", map[string]interface{}{
					"sabakanaddr": req.URL.String(),
					log.FnError:   err,
				})
				return nil, err
			}
			defer resp.Body.Close()
			err = json.NewDecoder(resp.Body).Decode(machines)
			if err != nil {
				return nil, err
			}
			return machines, nil
		}()
		if machines != nil {
			break
		}
	}
	return machines, err
}

func (c client) updateTargetEndpoints(machines *[]sabakan.Machine) error {
	services := c.k8s.CoreV1().Services(monitoringNamespace)
	_, err := services.Get(targetEndpointsName, metav1.GetOptions{})
	switch {
	case err == nil:
	case k8serrors.IsNotFound(err):
		_, err = services.Create(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: targetEndpointsName,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{Port: nodeExporterPort, TargetPort: intstr.FromInt(nodeExporterPort), Name: nodeExporterPortName},
					{Port: ckeEtcdPort, TargetPort: intstr.FromInt(ckeEtcdPort), Name: ckeEtcdPortName}},
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
		Addresses: make([]corev1.EndpointAddress, len(*machines)),
		Ports: []corev1.EndpointPort{
			{Port: nodeExporterPort, Name: nodeExporterPortName},
			{Port: ckeEtcdPort, Name: ckeEtcdPortName}},
	}
	for i, machine := range *machines {
		subset.Addresses[i].IP = machine.Spec.IPv4[0]
	}

	_, err = c.k8s.CoreV1().Endpoints(monitoringNamespace).Update(&corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetEndpointsName,
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
	serfc, err := serfclient.NewRPCClient("127.0.0.1:7373")
	if err != nil {
		log.ErrorExit(err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.ErrorExit(err)
	}

	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.ErrorExit(err)
	}
	client := client{http: localHTTPClient(), k8s: k8sClientSet, serf: serfc}

	members, err := client.GetMembers()
	if err != nil {
		log.ErrorExit(err)
	}

	bootservers := getBootServers(&members)

	machines, err := client.getMachinesFromSabakan(bootservers)
	if err != nil {
		log.ErrorExit(err)
	}

	err = client.updateTargetEndpoints(machines)
	if err != nil {
		log.ErrorExit(err)
	}
}
