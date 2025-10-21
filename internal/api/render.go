package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"mathsvg/internal/cache"
	"mathsvg/internal/renderer"
)

const (
	responseContentType = "image/svg+xml; charset=utf-8"
	errorSVG            = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 80"><rect width="100%" height="100%" fill="#fef2f2"/><text x="20" y="45" font-size="20" font-family="sans-serif" fill="#b91c1c">公式渲染失败，请检查输入</text></svg>`
)

// RenderHandler 负责对外暴露渲染 API
type RenderHandler struct {
	cache          *cache.Manager
	renderer       renderer.Renderer
	logger         *zap.Logger
	requestTimeout time.Duration
}

// NewRenderHandler 构建渲染处理器实例
func NewRenderHandler(cache *cache.Manager, renderer renderer.Renderer, logger *zap.Logger, timeout time.Duration) *RenderHandler {
	return &RenderHandler{
		cache:          cache,
		renderer:       renderer,
		logger:         logger,
		requestTimeout: timeout,
	}
}

// Register 将渲染接口挂载到指定的 Fiber 路由组
func (h *RenderHandler) Register(router fiber.Router) {
	router.Get("/render", h.handleRender)
}

// handleRender 为 GET /render 提供具体业务处理逻辑
func (h *RenderHandler) handleRender(c *fiber.Ctx) error {
	start := time.Now()
	tex := c.Query("tex")
	requestID := requestIDFromCtx(c)
	log := h.logger.With(zap.String("request_id", requestID))

	normalized, err := validateFormula(tex)
	if err != nil {
		status := classifyInputError(err)
		log.Warn("公式输入不合法", zap.Error(err))
		c.Set("Content-Type", responseContentType)
		return c.Status(status).SendString(errorSVG)
	}

	// 先生成缓存键，避免重复渲染
	cacheKey := hashFormula(normalized)
	reqCtx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// 一级缓存 → 二级缓存 → 缓存未命中时渲染
	svg, hitLevel := h.cache.Get(reqCtx, cacheKey)
	var renderDuration time.Duration
	if hitLevel == cache.HitNone {
		renderStart := time.Now()
		output, err := h.renderer.Render(normalized)
		renderDuration = time.Since(renderStart)
		if err != nil {
			// 若渲染失败，返回预置错误 SVG，避免前端渲染空白
			log.Error("渲染失败", zap.Error(err))
			c.Set("Content-Type", responseContentType)
			return c.Status(fiber.StatusUnprocessableEntity).SendString(errorSVG)
		}
		svg = output
		h.cache.Set(reqCtx, cacheKey, svg)
	}

	// 将关键指标写入结构化日志
	totalDuration := time.Since(start)
	log.Info("公式渲染完成",
		zap.Float64("request_duration_ms", float64(totalDuration.Microseconds())/1000.0),
		zap.Float64("render_duration_ms", float64(renderDuration.Microseconds())/1000.0),
		zap.String("cache_hit_level", string(hitLevel)),
		zap.Int("formula_length", len([]rune(normalized))),
	)

	c.Set("Content-Type", responseContentType)
	return c.SendString(svg)
}

// hashFormula 将公式内容转换为缓存键，减少重复计算
func hashFormula(tex string) string {
	sum := sha256.Sum256([]byte(tex))
	return hex.EncodeToString(sum[:])
}

func classifyInputError(err error) int {
	switch err {
	case ErrFormulaTooLarge:
		return fiber.StatusRequestEntityTooLarge
	case ErrInvalidCharacters:
		return fiber.StatusBadRequest
	case ErrEmptyFormula:
		return fiber.StatusBadRequest
	default:
		return fiber.StatusBadRequest
	}
}
