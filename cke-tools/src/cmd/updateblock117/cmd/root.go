package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "updateblock117",
		Short: "Update block device paths to upgrading from k8s 1.16 to 1.17 without draining Node.",
		Long: `updateblock117 is the program to manage block device files and its symlink paths 
created by Kubelet.
When upgrading Kubelet to k8s 1.17, we should fix the device paths which defined by Kubelet.
To operate this procedure by CKE, updateblock117 is used.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
