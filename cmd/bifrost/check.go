package main

import (
	"fmt"

	"github.com/farovictor/bifrost/config"
	"github.com/farovictor/bifrost/pkg/database"
	redis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check datastore connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbType := config.DBType()
		var dsn string
		if dbType == "postgres" || dbType == "sqlite" {
			dsn = config.DatabaseDSN()
			if dsn == "" {
				return fmt.Errorf("DATABASE_DSN is not set")
			}
		}
		db, err := database.Connect(dbType, dsn)
		if err != nil {
			return err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.Close()

		rdb := redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr(),
			Password: config.RedisPassword(),
			DB:       config.RedisDB(),
			Protocol: config.RedisProtocol(),
		})
		if err := rdb.Ping(cmd.Context()).Err(); err != nil {
			return err
		}

		fmt.Println("connections ok")
		return nil
	},
}
