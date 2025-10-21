package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"mathsvg/internal/cache"
	"mathsvg/internal/pkg/ctxkeys"
)

// HealthHandler 提供基础健康检查接口
type HealthHandler struct {
	cache   *cache.Manager
	logger  *zap.Logger
	started time.Time
}

// NewHealthHandler 构建健康检查处理器
func NewHealthHandler(cache *cache.Manager, logger *zap.Logger, started time.Time) *HealthHandler {
	return &HealthHandler{
		cache:   cache,
		logger:  logger,
		started: started,
	}
}

// Register 将健康检查接口挂载到路由上
func (h *HealthHandler) Register(router fiber.Router) {
	router.Get("/health", h.handleHealth)
}

func (h *HealthHandler) handleHealth(c *fiber.Ctx) error {
	requestID := requestIDFromCtx(c)
	stats := h.cache.Stats()

	response := fiber.Map{
		"status":     "ok",
		"uptime_ms":  time.Since(h.started).Milliseconds(),
		"request_id": requestID,
		"cache": fiber.Map{
			"local_entries": stats.LocalEntries,
			"hit_local":     stats.HitsLocal,
			"hit_redis":     stats.HitsRedis,
			"miss":          stats.Misses,
			"redis_enabled": stats.RedisEnabled,
			"redis_alive":   stats.RedisAlive,
		},
	}

	h.logger.Info("健康检查", zap.String("request_id", requestID))

	c.Set("Content-Type", "application/json; charset=utf-8")
	return c.Status(fiber.StatusOK).JSON(response)
}

func requestIDFromCtx(c *fiber.Ctx) string {
	if value := c.Locals(ctxkeys.RequestID); value != nil {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}
