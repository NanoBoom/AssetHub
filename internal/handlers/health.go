package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/NanoBoom/asethub/internal/cache"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *cache.RedisClient
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status   string `json:"status" example:"ok"`                // 整体状态: ok/degraded
	Database string `json:"database" example:"ok"`              // 数据库状态: ok/error
	Redis    string `json:"redis" example:"ok"`                 // Redis状态: ok/error
}

func NewHealthHandler(db *gorm.DB, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Check godoc
// @Summary      健康检查
// @Description  检查服务、数据库和 Redis 的健康状态
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "ok"
	if sqlDB, err := h.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		dbStatus = "error"
	}

	redisStatus := "ok"
	if err := h.redis.Ping(ctx); err != nil {
		redisStatus = "error"
	}

	status := "ok"
	if dbStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:   status,
		Database: dbStatus,
		Redis:    redisStatus,
	})
}
