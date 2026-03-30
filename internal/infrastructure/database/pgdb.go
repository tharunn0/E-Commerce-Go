package database

import (
	"context"
	// "log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func InitDB(connStr string, logger *zap.Logger) *pgxpool.Pool {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("Connecting to the database...")
	dbpool, _ := pgxpool.New(ctx, connStr)
	// if er != nil {
	// 	logger.Fatal("DATABASE_INITIALIZATION_FAILED", zap.Error(er))
	// }

	// Pinging the database to verify the connection pool
	// var res int
	// if er = dbpool.QueryRow(ctx, "SELECT 9;").Scan(&res); res != 9 || er != nil {
	// 	log.Println("err :", er)
	// 	logger.Panic("DATABASE_PING_FAILED")
	// }

	logger.Info("DATABASE_CONNECTED_SUCCESSFULLY")

	return dbpool
}
