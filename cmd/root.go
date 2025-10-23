package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vhost-manager",
	Short: "A tool for managing nginx virtual hosts",
	Long: `vhost-manager is a CLI tool that helps you manage nginx virtual hosts.
It can add, remove, and list virtual hosts by creating configuration files
in sites-available and creating symlinks in sites-enabled.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vhost-manager.yaml)")
	rootCmd.PersistentFlags().String("nginx-path", "/etc/nginx", "nginx configuration directory")
	rootCmd.PersistentFlags().Bool("dry-run", false, "show what would be done without making changes")

	// Bind flags to viper
	viper.BindPFlag("nginx.path", rootCmd.PersistentFlags().Lookup("nginx-path"))
	viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
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

		// Search config in home directory with name ".vhost-manager" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".vhost-manager")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}