package testhelper

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/taskflow_test?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping: cannot create db pool (%v). Set TEST_DB_URL or start a local test postgres.", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping: cannot connect to test db (%v). Set TEST_DB_URL or start a local test postgres.", err)
	}

	migrationsPath := repoMigrationsPath(t)
	if err := runMigrations(dbURL, migrationsPath); err != nil {
		pool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `
TRUNCATE TABLE tasks, projects, users RESTART IDENTITY CASCADE;
`)
		pool.Close()
	})

	return pool
}

func repoMigrationsPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	// tests/testhelper/setup.go -> repo root is ../..
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	return filepath.Join(root, "migrations")
}

func runMigrations(dbURL, migrationsDir string) error {
	m, err := migrate.New("file://"+migrationsDir, dbURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}
	return nil
}

