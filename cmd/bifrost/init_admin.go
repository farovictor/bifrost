package main

import (
	"fmt"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/database"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	"github.com/spf13/cobra"
)

var (
	initAdminName    string
	initAdminEmail   string
	initAdminOrgID   string
	initAdminOrgName string
)

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

		orgStore := orgs.NewPostgresStore(db)
		o := orgs.Organization{ID: initAdminOrgID, Name: initAdminOrgName}
		if err := orgStore.Create(o); err != nil && err != orgs.ErrOrgExists {
			return err
		}

		store := users.NewPostgresStore(db)
		key := config.AdminAPIKey()
		if key == "" {
			key = users.GenerateAPIKey()
		}
		u := users.User{
			Name:   initAdminName,
			Email:  initAdminEmail,
			APIKey: key,
		}
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
	initAdminCmd.Flags().StringVar(&initAdminName, "name", config.AdminName(), "admin user name")
	initAdminCmd.Flags().StringVar(&initAdminEmail, "email", config.AdminEmail(), "admin user email")
	initAdminCmd.Flags().StringVar(&initAdminOrgID, "org-id", config.AdminOrgID(), "admin organization id")
	initAdminCmd.Flags().StringVar(&initAdminOrgName, "org-name", config.AdminOrgName(), "admin organization name")
	rootCmd.AddCommand(initAdminCmd)
}
