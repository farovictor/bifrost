package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/FokusInternal/bifrost/pkg/rootkeys"
	"github.com/spf13/cobra"
)

var (
	rootKeyUpdateID     string
	rootKeyUpdateAPIKey string
)

var rootKeyUpdateCmd = &cobra.Command{
	Use:   "rootkey-update",
	Short: "Update a root key",
	RunE: func(cmd *cobra.Command, args []string) error {
		rk := rootkeys.RootKey{ID: rootKeyUpdateID, APIKey: rootKeyUpdateAPIKey}
		body, err := json.Marshal(rk)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPut, serverAddr+"/v1/rootkeys/"+rootKeyUpdateID, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server error: %s", bytes.TrimSpace(b))
		}
		io.Copy(os.Stdout, resp.Body)
		return nil
	},
}

func init() {
	rootKeyUpdateCmd.Flags().StringVar(&rootKeyUpdateID, "id", "", "root key id")
	rootKeyUpdateCmd.Flags().StringVar(&rootKeyUpdateAPIKey, "apikey", "", "API key")
	rootKeyUpdateCmd.MarkFlagRequired("id")
	rootKeyUpdateCmd.MarkFlagRequired("apikey")
	rootCmd.AddCommand(rootKeyUpdateCmd)
}
