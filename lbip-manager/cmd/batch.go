/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"log/slog"

	"github.com/cybozu/neco-containers/lbip-manager/pkg/k8s"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Run batch job to assign IP addresses",
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(context.Background()); err != nil {
			log.Error("batch command failed", slog.Any("error", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(ctx context.Context) error {
	svcs, err := k8s.ListAllServices(ctx)
	if err != nil {
		return err
	}

	for _, svc := range svcs.Items {
		if !isTargetLB(&svc) {
			continue
		}

		if err := fillIP(ctx, &svc); err != nil {
			log.Error("failed to fill IP address",
				slog.Any("namespace", svc.Namespace),
				slog.Any("service", svc.Name),
				slog.Any("error", err))
		}
	}
	return nil
}

func isTargetLB(svc *v1.Service) bool {
	log := log.With("namespace", svc.Namespace)
	log = log.With("service", svc.Name)

	if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
		log.Info("skip because service is not LoadBalancer", slog.Any("type", svc.Spec.Type))
		return false
	}

	if svc.Spec.LoadBalancerIP != "" {
		log.Info("skip because service already has spec.loadBalancerIP", slog.Any("ip", svc.Spec.LoadBalancerIP))
		return false
	}

	if svc.Status.LoadBalancer.Ingress == nil {
		log.Info("skip because service does not have LoadBalancerIP yet")
		return false
	}

	if len(svc.Status.LoadBalancer.Ingress) != 1 {
		log.Info("skip because service has multiple LoadBalancerIPs")
		return false
	}

	return true
}

func fillIP(ctx context.Context, svc *v1.Service) error {
	eip := svc.Status.LoadBalancer.Ingress[0].IP
	return k8s.ApplyLoadBalancerIP(ctx, svc, eip)

}
