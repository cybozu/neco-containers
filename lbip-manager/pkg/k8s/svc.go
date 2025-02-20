package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1apply "k8s.io/client-go/applyconfigurations/core/v1"
)

func ListAllServices(ctx context.Context) (*v1.ServiceList, error) {
	return client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
}

func ApplyLoadBalancerIP(ctx context.Context, svc *v1.Service, ip string) error {
	s := v1apply.Service(svc.Name, svc.Namespace)
	s.WithSpec(
		v1apply.ServiceSpec().WithLoadBalancerIP(ip),
	)

	opts := metav1.ApplyOptions{
		Force:        true,
		FieldManager: FieldManagerName,
	}

	_, err := client.CoreV1().Services("").Apply(ctx, s, opts)
	return err
}
