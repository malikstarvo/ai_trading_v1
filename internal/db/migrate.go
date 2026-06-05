package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.FS) error {
	if err := ensureSchemaTable(ctx, pool); err != nil {
		return fmt.Errorf("ensure schema table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		if err := applyMigration(ctx, pool, migrationsFS, name); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func ensureSchemaTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename    TEXT PRIMARY KEY,
			checksum    TEXT NOT NULL,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.FS, name string) error {
	var existingChecksum *string
	err := pool.QueryRow(ctx, `SELECT checksum FROM schema_migrations WHERE filename = $1`, name).Scan(&existingChecksum)
	if err != nil && err.Error() != "no rows in result set" {
		return fmt.Errorf("check migration %s: %w", name, err)
	}

	content, err := fs.ReadFile(migrationsFS, name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(content))

	if existingChecksum != nil {
		if *existingChecksum != checksum {
			log.Fatalf("MIGRATION MISMATCH: %s has different checksum than applied version. "+
				"Expected %s, got %s. This migration cannot be changed after deployment.", name, *existingChecksum, checksum)
		}
		return nil
	}

	if _, err := pool.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("execute %s: %w", name, err)
	}

	if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)`, name, checksum); err != nil {
		return fmt.Errorf("record %s: %w", name, err)
	}

	log.Printf("migration applied: %s", name)
	return nil
}
