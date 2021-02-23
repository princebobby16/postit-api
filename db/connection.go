package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"postit-backend-api/pkg/logs"
)

var Connection *sql.DB

func Connect() {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	Connection = db

	err = db.Ping()
	if err != nil {
		log.Printf("Unable to connect to database")
		panic(err)
	}

	logs.Logger.Info("Connected to Postgres DB successfully")
}

func Disconnect() {
	logs.Logger.Info("Attempting to disconnect from db....")
	err := Connection.Close()
	if err != nil {
		logs.Logger.Error(err)
	}
	logs.Logger.Info("Disconnected from db successfully...")
}
