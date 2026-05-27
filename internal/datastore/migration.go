package datastore

import (
	"context"
	"database/sql"
	"embed"
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

func newVectorTableMigration(_ *conf.Config) *goose.Migration {
	up := &goose.GoFunc{
		RunTx: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS embedding (
				rowid INTEGER PRIMARY KEY,
				embedding BLOB NOT NULL
			)`)
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
