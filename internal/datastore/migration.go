package datastore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/niedch/vectree/internal/conf"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB, cfg *conf.Config) error {
	goose.SetLogger(goose.NopLogger())

	migrationFS, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(
		"sqlite3",
		db,
		migrationFS,
		goose.WithGoMigrations(newVectorTableMigration(cfg)),
	)
	if err != nil {
		return err
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return err
	}

	return nil
}

func newVectorTableMigration(cfg *conf.Config) *goose.Migration {
	vertexSize := cfg.AI.VertexSize
	if vertexSize == 0 {
		vertexSize = conf.DEFAULT_VERTEX_SIZE
	}

	up := &goose.GoFunc{
		RunTx: func(ctx context.Context, tx *sql.Tx) error {
			query := fmt.Sprintf(
				"CREATE VIRTUAL TABLE IF NOT EXISTS embedding USING vec0(embedding FLOAT[%d])",
				vertexSize,
			)
			_, err := tx.ExecContext(ctx, query)
			return err
		},
	}

	down := &goose.GoFunc{
		RunTx: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS embedding")
			return err
		},
	}

	return goose.NewGoMigration(20251114113611, up, down)
}
