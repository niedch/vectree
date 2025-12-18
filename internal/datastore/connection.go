package datastore

import (
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"broadcom.com/vertex-ingestor/internal/conf"
)

func OpenConnection(config *conf.Config) (*sqlx.DB, error) {
	// Load sqlite-vec extension
	sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", config.Database.ConnectionString)
	if err != nil {
		return nil, err
	}

	RunMigrations(db.Unsafe().DB)
	return db, nil
}
