package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/spf13/cobra"
)

var (
	rootKeyAddID     string
	rootKeyAddAPIKey string
)

var rootKeyAddCmd = &cobra.Command{
	Use:   "rootkey-add",
	Short: "Add a root key",
	RunE: func(cmd *cobra.Command, args []string) error {
		rk := rootkeys.RootKey{ID: rootKeyAddID, APIKey: rootKeyAddAPIKey}
		body, err := json.Marshal(rk)
		if err != nil {
			return err
		}
		resp, err := http.Post(serverAddr+"/v1/rootkeys", "application/json", bytes.NewReader(body))
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
	rootKeyAddCmd.Flags().StringVar(&rootKeyAddID, "id", "", "root key id")
	rootKeyAddCmd.Flags().StringVar(&rootKeyAddAPIKey, "apikey", "", "API key")
	rootKeyAddCmd.MarkFlagRequired("id")
	rootKeyAddCmd.MarkFlagRequired("apikey")
	rootCmd.AddCommand(rootKeyAddCmd)
}
