package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bifrost",
	Short: "Bifrost command line interface",
}

var serverAddr string

func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "addr", "http://localhost:3333", "bifrost API address")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
