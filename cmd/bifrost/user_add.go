package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	userAddID      string
	userAddOrgID   string
	userAddOrgName string
	userAddRole    string
)

var userAddCmd = &cobra.Command{
	Use:   "user-add",
	Short: "Create a user and optionally register an organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]string{
			"id":       userAddID,
			"org_id":   userAddOrgID,
			"org_name": userAddOrgName,
			"role":     userAddRole,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		resp, err := http.Post(serverAddr+"/v1/users", "application/json", bytes.NewReader(body))
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
	userAddCmd.Flags().StringVar(&userAddID, "id", "", "user id")
	userAddCmd.Flags().StringVar(&userAddOrgID, "org-id", "", "organization id")
	userAddCmd.Flags().StringVar(&userAddOrgName, "org-name", "", "organization name")
	userAddCmd.Flags().StringVar(&userAddRole, "role", "", "membership role")
	userAddCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(userAddCmd)
}
