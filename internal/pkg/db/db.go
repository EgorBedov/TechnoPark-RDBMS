package db

import (
	"github.com/jackc/pgx"
	"log"
	"sync"
)

var (
	// TODO: add it as environment variable
	host     = "localhost"
	port     = uint16(5432)
	user     = "docker"
	password = "docker"
	dbname   = "docker"
)


var db *pgx.ConnPool = nil
var syncOnce = sync.Once{}

func ConnectToDB() *pgx.ConnPool {
	syncOnce.Do(func() {
		pgxConfig := pgx.ConnConfig{
			Host:     host,
			Port:     port,
			Database: dbname,
			User:     user,
			Password: password,
		}
		pgxConnPoolConfig := pgx.ConnPoolConfig{
			MaxConnections: 1,
			ConnConfig: pgxConfig,
		}
		dbase, err := pgx.NewConnPool(pgxConnPoolConfig)
		if err != nil {
			db = nil
		} else {
			db = dbase
		}
	})
	log.Println("Connected to db: name=", dbname, ", user=", user, ", port=", port)
	return db
}
