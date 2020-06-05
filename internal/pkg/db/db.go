package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"sync"
)

const (
	Host     = "localhost"
	Port     = uint16(5432)
	User     = "docker"
	Password = "docker"
	dbName   = "docker"
)


var db *pgxpool.Pool = nil
var syncOnce = sync.Once{}

func ConnectToDB() *pgxpool.Pool {
	syncOnce.Do(func() {
		config, err := pgxpool.ParseConfig(fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v", User, Password, Host, Port, dbName))
		if err != nil {
			panic("Failed to parse config")
		}
		config.MaxConns = 100
		config.MinConns = 10
		config.ConnConfig.LogLevel = pgx.LogLevelDebug
		pconf, err := pgxpool.ConnectConfig(context.Background(), config)
		if err != nil {
			db = nil
			fmt.Println(err)
		} else {
			db = pconf
		}
	})
	return db
}

func InsertPlaceholders(query string, length int) string {
	var indices []interface{}
	for iii := 0; iii < length; iii++ {
		indices = append(indices, iii + 1)
	}
	return fmt.Sprintf(query, indices...)
}
