package datastore

import (
	"database/sql"
	"fmt"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"

	"broadcom.com/vertex-ingestor/internal/conf"
)

func OpenConnection(config *conf.Config) (Queries, error) {
	// Load sqlite-vec extension
	sqlite_vec.Auto()

	dbPath := fmt.Sprintf("%s?cache=shared&mode=rw", config.DATABASE_NAME)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return Queries{}, err
	}

	RunMigrations(db)
	return *New(db), nil
}
