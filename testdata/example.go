package testdata

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

//go:generate embed file -var createDbSQL --source create_query.sql
var createDbSQL string

func CreateDB(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	_, err = db.Exec(createDbSQL)
	return err
}
