package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var rootKeyDeleteCmd = &cobra.Command{
	Use:   "rootkey-delete [id]",
	Short: "Delete a root key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, err := http.NewRequest(http.MethodDelete, serverAddr+"/v1/rootkeys/"+args[0], nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server error: %s", bytes.TrimSpace(b))
		}
		fmt.Println("deleted")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rootKeyDeleteCmd)
}
