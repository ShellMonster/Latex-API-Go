package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mathsvg/internal/api"
	"mathsvg/internal/cache"
	"mathsvg/internal/config"
	"mathsvg/internal/logging"
	"mathsvg/internal/renderer"
	"mathsvg/internal/server"

	"go.uber.org/zap"
)

func main() {
	bootTime := time.Now()

	// 加载配置，确保不同环境都能读取统一的服务参数
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// 初始化结构化日志，便于后续排查性能与业务问题
	logger, err := logging.New(cfg.Log)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	// 准备缓存组件，先构建 BigCache，再视情况启用 Redis
	cacheManager, err := cache.NewManager(cfg.Cache, logger)
	if err != nil {
		logger.Fatal("缓存初始化失败", zap.Error(err))
	}
	defer func() { _ = cacheManager.Close() }()

	// 优先尝试加载 Rust 渲染器，如失败则降级为占位实现
	rendererImpl, err := renderer.NewFFIRenderer()
	if err != nil {
		logger.Warn("Rust 渲染器初始化失败，降级为占位实现", zap.Error(err))
		rendererImpl = renderer.NewStub()
	} else {
		logger.Info("Rust 渲染器初始化成功，启用真实渲染")
	}

	// 将渲染逻辑封装到统一的 Handler 中，方便后续扩展监控与鉴权
	renderHandler := api.NewRenderHandler(cacheManager, rendererImpl, logger, cfg.Server.RequestTimeout)
	healthHandler := api.NewHealthHandler(cacheManager, logger, bootTime)

	// 构建 HTTP 服务，里面会自动挂载路由、中间件等组件
	httpServer := server.NewHTTPServer(cfg.Server, logger, renderHandler, healthHandler)

	// 采用独立协程启动服务，主协程负责监听退出信号
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	// 捕获系统信号，实现优雅停机
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info("接收到停止信号，开始执行优雅停机流程")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("优雅停机失败", zap.Error(err))
		return
	}

	// 预留时间给后台任务收尾
	time.Sleep(200 * time.Millisecond)
	logger.Info("服务已安全退出")
}
