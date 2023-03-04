package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information for terraformify",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CLI version: %s\n", getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
