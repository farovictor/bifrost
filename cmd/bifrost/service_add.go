package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/FokusInternal/bifrost/pkg/services"
	"github.com/spf13/cobra"
)

var (
	serviceAddID       string
	serviceAddEndpoint string
	serviceAddAPIKey   string
)

var serviceAddCmd = &cobra.Command{
	Use:   "service-add",
	Short: "Add a service",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := services.Service{ID: serviceAddID, Endpoint: serviceAddEndpoint, APIKey: serviceAddAPIKey}
		body, err := json.Marshal(svc)
		if err != nil {
			return err
		}
		resp, err := http.Post(serverAddr+"/v1/services", "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server error: %s", bytes.TrimSpace(b))
		}
		io.Copy(os.Stdout, resp.Body)
		return nil
	},
}

func init() {
	serviceAddCmd.Flags().StringVar(&serviceAddID, "id", "", "service id")
	serviceAddCmd.Flags().StringVar(&serviceAddEndpoint, "endpoint", "", "service endpoint")
	serviceAddCmd.Flags().StringVar(&serviceAddAPIKey, "apikey", "", "API key")
	serviceAddCmd.MarkFlagRequired("id")
	serviceAddCmd.MarkFlagRequired("endpoint")
	serviceAddCmd.MarkFlagRequired("apikey")
	rootCmd.AddCommand(serviceAddCmd)
}
