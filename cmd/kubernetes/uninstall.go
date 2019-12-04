package main

import (
	"get.porter.sh/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(mixin *kubernetes.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Use kubectl to delete resources contained in a manifest from a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Uninstall()
		},
	}
}
