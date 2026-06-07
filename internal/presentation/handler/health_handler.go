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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	dbStatus := gin.H{
		"status": "up",
	}

	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("Database ping failed", zap.Error(err))

		dbStatus["status"] = "down"
		dbStatus["error"] = err.Error()
	}

	redisStatus := gin.H{
		"status": "up",
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.Error("Redis ping failed", zap.Error(err))

		redisStatus["status"] = "down"
		redisStatus["error"] = err.Error()
	}

	httpStatus := http.StatusOK
	if dbStatus["status"] == "down" || redisStatus["status"] == "down" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":  "success",
		"message": "Health check completed",
		"data": gin.H{
			"server": gin.H{
				"status": "up",
			},
			"database":  dbStatus,
			"redis":     redisStatus,
			"timestamp": time.Now().UTC(),
		},
	})
}
