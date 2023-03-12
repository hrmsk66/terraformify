package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var version string

func getVersion() string {
	if version != "" {
		return version
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}

	return "(unknown)"
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terraformify",
	Short: "A CLI that generates TF files to manage existing Fastly services with Terraform",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terraformify.yaml)")
	rootCmd.PersistentFlags().StringP("working-dir", "d", ".", "Terraform working directory")
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "Interactively select associated resources to import")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "Fastly API token (or via FASTLY_API_KEY)")
	rootCmd.PersistentFlags().BoolP("skip-edit-state", "s", false, "Skip editing terraform.tfstate and leave it untouched (Note: Diffs will be detected on terraform plan/apply)")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "Answer yes automatically to all Yes/No confirmations")

	// Associate --token with the env ver, FASTLY_API_KEY
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("FASTLY")
	if err := viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key")); err != nil {
		log.Fatal(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".terraformify" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".terraformify")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
