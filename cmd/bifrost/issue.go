package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/FokusInternal/bifrost/pkg/keys"
	"github.com/spf13/cobra"
)

var (
	issueID     string
	issueScope  string
	issueTarget string
	issueTTL    time.Duration
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue a virtual key",
	RunE: func(cmd *cobra.Command, args []string) error {
		k := keys.VirtualKey{
			ID:        issueID,
			Scope:     issueScope,
			Target:    issueTarget,
			ExpiresAt: time.Now().Add(issueTTL),
		}
		body, err := json.Marshal(k)
		if err != nil {
			return err
		}
		resp, err := http.Post(serverAddr+"/v1/keys", "application/json", bytes.NewReader(body))
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
	issueCmd.Flags().StringVar(&issueID, "id", "", "ID of the key")
	issueCmd.Flags().StringVar(&issueScope, "scope", "", "scope for the key")
	issueCmd.Flags().StringVar(&issueTarget, "target", "", "target service")
	issueCmd.Flags().DurationVar(&issueTTL, "ttl", time.Hour, "time to live")
	issueCmd.MarkFlagRequired("id")
	issueCmd.MarkFlagRequired("target")
	rootCmd.AddCommand(issueCmd)
}
