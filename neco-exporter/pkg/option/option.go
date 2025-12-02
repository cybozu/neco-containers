package option

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
)

var (
	Scope          string
	Port           int
	CollectorNames []string
	Interval       time.Duration

	ControllerMetricsPort   int
	ControllerProbePort     int
	LeaderElectionNamespace string
)

func SetupOptionFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Scope, "scope", constants.ScopeCluster, fmt.Sprintf("Collection scope (%s or %s)", constants.ScopeCluster, constants.ScopeNode))
	cmd.Flags().IntVar(&Port, "port", 8080, "Specify port to expose metrics")
	cmd.Flags().StringSliceVar(&CollectorNames, "collectors", []string{"bpf"}, "Specify collectors to activate")
	cmd.Flags().DurationVar(&Interval, "interval", time.Second*30, "Interval to update metrics")
}
