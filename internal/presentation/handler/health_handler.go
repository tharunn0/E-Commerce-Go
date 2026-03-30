package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	logger *zap.Logger
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check Database
	dbStatus := "up"
	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("Database ping failed", zap.Error(err))
		dbStatus = "down"
	}

	// Check Redis
	redisStatus := "up"
	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.Error("Redis ping failed", zap.Error(err))
		redisStatus = "down"
	}

	status := http.StatusOK
	if dbStatus == "down" || redisStatus == "down" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": "success",
		"data": gin.H{
			"server":    "up",
			"database":  dbStatus,
			"redis":     redisStatus,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}
