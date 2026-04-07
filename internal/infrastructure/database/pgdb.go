package database

import (
	"context"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Driver for Postgres
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tharunn0/E-Commerce-Go/migration"
	"go.uber.org/zap"
)

func InitDB(connStr string, logger *zap.Logger) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Create the Pool
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		logger.Fatal("DATABASE_CONNECTION_FAILED", zap.Error(err))
	}

	// 2. Verify connection
	if err := dbpool.Ping(ctx); err != nil {
		logger.Fatal("DATABASE_PING_FAILED", zap.Error(err))
	}

	// 3. Run Migrations
	runMigrations(connStr, logger)

	logger.Info("DATABASE_READY_WITH_MIGRATIONS")
	return dbpool
}

func runMigrations(connStr string, logger *zap.Logger) {

	d, err := iofs.New(migration.FS, ".")
	if err != nil {
		logger.Fatal("MIGRATION_INIT_FAILED", zap.Error(err))
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		logger.Fatal("MIGRATION_INIT_FAILED", zap.Error(err))
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("MIGRATIONS_ALREADY_UP_TO_DATE")
		} else {
			logger.Fatal("MIGRATION_FAILED", zap.Error(err))
		}
	} else {
		logger.Info("MIGRATIONS_APPLIED_SUCCESSFULLY")
	}
}
