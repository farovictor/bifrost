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
		o, err := orgStore.GetByName(initAdminOrgName)
		if err != nil {
			if err != orgs.ErrOrgNotFound {
				return err
			}
			o = orgs.Organization{ID: utils.GenerateID(), Name: initAdminOrgName, Domain: initAdminOrgDomain, Email: initAdminOrgEmail}
			if err := orgStore.Create(o); err != nil {
				return err
			}
		}

		store := users.NewPostgresStore(db)
		key := config.AdminAPIKey()
		if key == "" {
			key = users.GenerateAPIKey()
		}
		u, err := store.GetByEmail(initAdminEmail)
		userExists := true
		if err != nil {
			if err != users.ErrUserNotFound {
				return err
			}
			userExists = false
			u = users.User{
				ID:     utils.GenerateID(),
				Name:   initAdminName,
				Email:  initAdminEmail,
				APIKey: key,
			}
			if err := store.Create(u); err != nil {
				return err
			}
		}
		memStore := orgs.NewPostgresMembershipStore(db)
		_, err = memStore.Get(u.ID, o.ID)
		membershipExists := true
		if err != nil {
			if err != orgs.ErrMembershipNotFound {
				return err
			}
			membershipExists = false
			m := orgs.Membership{UserID: u.ID, OrgID: o.ID, Role: initAdminRole}
			if err := memStore.Create(m); err != nil {
				return err
			}
		}
		if userExists && membershipExists {
			fmt.Fprintln(cmd.ErrOrStderr(), "warning: user and membership already exist")
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
