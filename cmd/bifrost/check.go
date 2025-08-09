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
		if dsn := config.PostgresDSN(); dsn != "" {
			db, err := database.Connect(config.DBType(), dsn)
			if err != nil {
				return err
			}
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			sqlDB.Close()
		}

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
