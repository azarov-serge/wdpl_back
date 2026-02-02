package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB — тонкая обёртка над *sql.DB, чтобы при необходимости заменить реализацию.
type DB struct {
	*sql.DB
}

func Connect(cfg *config.Config, log logger.Logger) (*DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info("connected to database")

	return &DB{DB: db}, nil
}

// RunMigrations применяет простые SQL‑миграции из каталога migrations.
// Важно: миграции должны быть идемпотентными (IF NOT EXISTS и т.п.).
func RunMigrations(db *DB, log logger.Logger) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		// Если каталог не найден — просто пропускаем (полезно на ранней стадии разработки).
		if strings.Contains(err.Error(), "file does not exist") {
			log.Info("no migrations directory found, skipping")
			return nil
		}
		return fmt.Errorf("read migrations dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migrations tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := "migrations/" + e.Name()
		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", e.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("exec migration %s: %w", e.Name(), err)
		}
		log.Info("migration applied", "file", e.Name())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations tx: %w", err)
	}

	return nil
}
