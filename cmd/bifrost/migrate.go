package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/farovictor/bifrost/config"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		dsn := config.PostgresDSN()
		if dsn == "" {
			return fmt.Errorf("POSTGRES_DSN is not set")
		}
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			return err
		}
		files, err := filepath.Glob(filepath.Join("migrations", "*.sql"))
		if err != nil {
			return err
		}
		sort.Strings(files)
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			if _, err := db.Exec(string(b)); err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "applied %d migrations\n", len(files))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
