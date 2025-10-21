package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"mathsvg/internal/api"
	"mathsvg/internal/config"
	"mathsvg/internal/pkg/ctxkeys"
)

// HTTPServer 封装 Fiber 实例，统一管理中间件、路由等配置
type HTTPServer struct {
	app    *fiber.App
	cfg    config.Server
	logger *zap.Logger
}

// NewHTTPServer 根据配置创建服务，同时注册 API
func NewHTTPServer(cfg config.Server, logger *zap.Logger, renderHandler *api.RenderHandler, healthHandler *api.HealthHandler) *HTTPServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Prefork:               cfg.Prefork,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		IdleTimeout:           cfg.IdleTimeout,
		BodyLimit:             cfg.MaxRequestBodyMB * 1024 * 1024,
		ServerHeader:          "MathSVG-Go",
	})

	app.Use(func(c *fiber.Ctx) error {
		requestID := uuid.NewString()
		c.Set("X-Request-ID", requestID)
		c.Locals(ctxkeys.RequestID, requestID)
		return c.Next()
	})

	app.Use(recover.New())
	if cfg.EnableCompression {
		app.Use(compress.New())
	}

	// 同时兼容 /render 与 /api/v1/render 两种路径，方便后续网关或版本化
	if healthHandler != nil {
		healthHandler.Register(app)
	}
	renderHandler.Register(app)
	apiGroup := app.Group("/api/v1")
	renderHandler.Register(apiGroup)

	return &HTTPServer{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
}

// Start 启动 HTTP 服务
func (s *HTTPServer) Start() error {
	s.logger.Info("HTTP 服务启动", zap.String("listen", s.cfg.Address))
	return s.app.Listen(s.cfg.Address)
}

// Shutdown 实现优雅下线
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
