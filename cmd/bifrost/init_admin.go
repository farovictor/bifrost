package main

import (
	"fmt"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/database"
	"github.com/farovictor/bifrost/pkg/users"
	"github.com/spf13/cobra"
)

var initAdminID string

var initAdminCmd = &cobra.Command{
	Use:   "init-admin",
	Short: "Create an admin user in the database",
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

		store := users.NewPostgresStore(db)
		key := config.AdminAPIKey()
		if key == "" {
			key = users.GenerateAPIKey()
		}
		u := users.User{ID: initAdminID, APIKey: key}
		if err := store.Create(u); err != nil {
			if err == users.ErrUserExists {
				return fmt.Errorf("user already exists")
			}
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), u.APIKey)
		return nil
	},
}

func init() {
	initAdminCmd.Flags().StringVar(&initAdminID, "id", config.AdminID(), "admin user id")
	rootCmd.AddCommand(initAdminCmd)
}
