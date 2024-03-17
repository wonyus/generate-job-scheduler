package mssql

import (
	"database/sql"
	"os"

	"github.com/wonyus/generate-job-scheduler/utils"
)

var (
	dbUrl = utils.Strip(os.Getenv("DB_URL"))
)

func NewDefaultClient(dbName string) (*sql.DB, error) {
	sqlDB, err := sql.Open("postgres", dbUrl+dbName)
	if err != nil {
		panic(err)
	}

	err = sqlDB.Ping()
	if err != nil {
		panic(err)
	}

	return sqlDB, nil
}
