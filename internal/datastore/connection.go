package datastore

import (
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/niedch/vectree/internal/conf"
)

func OpenConnection(config *conf.Config) (*sqlx.DB, error) {
	// Load sqlite-vec extension
	sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", config.Database.ConnectionString)
	if err != nil {
		return nil, err
	}

	vertexSize := config.AI.VertexSize
	if vertexSize == 0 {
		vertexSize = conf.DEFAULT_VERTEX_SIZE
	}

	err = RunMigrations(db.Unsafe().DB, config)
	return db, err
}
