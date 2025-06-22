package main

import (
	"database/sql"
	"fmt"

	"github.com/farovictor/bifrost/config"
	_ "github.com/lib/pq"
	redis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check datastore connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dsn := config.PostgresDSN(); dsn != "" {
			db, err := sql.Open("postgres", dsn)
			if err != nil {
				return err
			}
			if err := db.Ping(); err != nil {
				return err
			}
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
