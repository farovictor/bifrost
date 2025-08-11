package main

import (
	"fmt"
	"time"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/auth"
	"github.com/farovictor/bifrost/pkg/database"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/users"
	"github.com/farovictor/bifrost/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	initAdminName      string
	initAdminEmail     string
	initAdminOrgName   string
	initAdminOrgEmail  string
	initAdminOrgDomain string
	initAdminRole      string
)

var initAdminCmd = &cobra.Command{
	Use:   "init-admin",
	Short: "Create an admin user in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbType := config.DBType()
		dsn := config.DatabaseDSN()
		if dbType == "postgres" && dsn == "" {
			return fmt.Errorf("DATABASE_DSN is not set")
		}
		db, err := database.Connect(dbType, dsn)
		if err != nil {
			return err
		}
		sqlDB, _ := db.DB()
		defer sqlDB.Close()

		orgStore := orgs.NewSQLStore(db)
		orgID := utils.GenerateID()
		o := orgs.Organization{ID: orgID, Name: initAdminOrgName, Domain: initAdminOrgDomain, Email: initAdminOrgEmail}
		if err := orgStore.Create(o); err != nil {
			return err
		}

		store := users.NewSQLStore(db)
		key := config.AdminAPIKey()
		if key == "" {
			key = users.GenerateAPIKey()
		}
		userID := utils.GenerateID()
		u := users.User{
			ID:     userID,
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
		memStore := orgs.NewSQLMembershipStore(db)
		m := orgs.Membership{UserID: u.ID, OrgID: o.ID, Role: initAdminRole}
		if err := memStore.Create(m); err != nil {
			return err
		}
		tok, err := auth.Sign(auth.AuthToken{
			UserID:    u.ID,
			OrgID:     o.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), tok)
		// fmt.Fprintln(cmd.OutOrStdout(), u.APIKey)
		return nil
	},
}

func init() {
	initAdminCmd.Flags().StringVar(&initAdminName, "name", config.AdminName(), "admin user name")
	initAdminCmd.Flags().StringVar(&initAdminEmail, "email", config.AdminEmail(), "admin user email")
	initAdminCmd.Flags().StringVar(&initAdminOrgName, "org-name", config.AdminOrgName(), "admin organization name")
	initAdminCmd.Flags().StringVar(&initAdminOrgEmail, "org-email", config.AdminOrgEmail(), "admin organization email")
	initAdminCmd.Flags().StringVar(&initAdminOrgDomain, "org-domain", config.AdminOrgDomain(), "admin organization domain")
	initAdminCmd.Flags().StringVar(&initAdminRole, "role", config.AdminRole(), "admin membership role")
	rootCmd.AddCommand(initAdminCmd)
}
