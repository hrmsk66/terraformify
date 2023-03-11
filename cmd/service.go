package cmd

import (
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use: "service",
}

func init() {
	rootCmd.AddCommand(serviceCmd)

	// Persistent flags
	serviceCmd.PersistentFlags().StringP("resource-name", "n", "service", "Target Terraform resource name")
	serviceCmd.PersistentFlags().IntP("version", "v", 0, "Version of the service to be imported")
	serviceCmd.PersistentFlags().BoolP("manage-all", "m", false, "Manage all associated resources")
	serviceCmd.PersistentFlags().BoolP("force-destroy", "f", false, "Set force-destroy to true for the service and associated resources")
}
