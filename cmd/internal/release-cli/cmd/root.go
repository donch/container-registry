package cmd

import (
	"os"

	"github.com/docker/distribution/cmd/internal/release-cli/configuration"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "release",
	Short: "A CLI tool for Container Registry releases",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile string
var tag string
var authToken string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.config.yaml)")

	rootCmd.PersistentFlags().StringVar(&tag, "tag", "", "Release version")
	rootCmd.MarkPersistentFlagRequired("tag")

	rootCmd.PersistentFlags().StringVar(&authToken, "token", "", "Trigger token or an auth token of the project to release to")
	rootCmd.MarkPersistentFlagRequired("token")
}

func initConfig() {
	if cfgFile != "" {
		configuration.SetConfig(cfgFile)
	} else {
		configuration.InitConfig(tag, authToken)
	}
}
