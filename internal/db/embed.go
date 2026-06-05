package db

import (
	"context"
	"embed"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunAllMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	subFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	return RunMigrations(ctx, pool, subFS)
}
