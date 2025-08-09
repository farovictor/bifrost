package main

import (
	"os"

	"github.com/farovictor/bifrost/config"
	"github.com/spf13/cobra"
)

var (
	dbType     string
	mode       string
	serverAddr string
)

var rootCmd = &cobra.Command{
	Use:   "bifrost",
	Short: "Bifrost command line interface",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if dbType != "" {
			os.Setenv("BIFROST_DB", dbType)
		}
		if mode != "" {
			os.Setenv("BIFROST_MODE", mode)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "addr", "http://localhost:3333", "bifrost API address")
	rootCmd.PersistentFlags().StringVar(
		&dbType,
		"db",
		config.DBType(),
		"database backend to use (sqlite or postgres). Flag takes precedence over BIFROST_DB",
	)
	rootCmd.PersistentFlags().StringVar(
		&mode,
		"mode",
		config.Mode(),
		"application mode. Flag takes precedence over BIFROST_MODE",
	)
	rootCmd.AddCommand(checkCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
