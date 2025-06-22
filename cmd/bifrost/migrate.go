package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/database"
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
		db, err := database.Connect(dsn)
		if err != nil {
			return err
		}
		sqlDB, _ := db.DB()
		defer sqlDB.Close()
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
			if err := db.Exec(string(b)).Error; err != nil {
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
