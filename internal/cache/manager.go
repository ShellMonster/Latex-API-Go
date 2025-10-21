package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	bigcache "github.com/allegro/bigcache/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"mathsvg/internal/config"
)

// HitLevel 用中文记录缓存命中的层级，方便日志分析
type HitLevel string

const (
	HitNone  HitLevel = "miss"  // 未命中任何缓存
	HitLocal HitLevel = "local" // 命中一级缓存
	HitRedis HitLevel = "redis" // 命中二级缓存
)

// Manager 统一管理一级 BigCache 与二级 Redis 的协同
type Manager struct {
	local        *bigcache.BigCache
	redis        *redis.Client
	redisEnabled bool
	redisTTL     time.Duration
	logger       *zap.Logger

	hitsLocal atomic.Uint64
	hitsRedis atomic.Uint64
	misses    atomic.Uint64

	redisAlive atomic.Bool
}

// Stats 描述缓存的关键运行指标
type Stats struct {
	LocalEntries int
	HitsLocal    uint64
	HitsRedis    uint64
	Misses       uint64
	RedisEnabled bool
	RedisAlive   bool
}

// NewManager 根据配置初始化缓存组件
func NewManager(cfg config.Cache, logger *zap.Logger) (*Manager, error) {
	localCache, err := bigcache.NewBigCache(bigcache.Config{
		Shards:             1024,
		LifeWindow:         cfg.LocalLifeWindow,
		CleanWindow:        cfg.LocalCleanWindow,
		MaxEntriesInWindow: 100_000,
		MaxEntrySize:       1024,
		Verbose:            false,
		HardMaxCacheSize:   cfg.LocalHardMaxCacheMB,
		StatsEnabled:       true,
	})
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		local:        localCache,
		redisEnabled: cfg.RedisEnabled,
		redisTTL:     cfg.RedisTTL,
		logger:       logger,
	}

	if cfg.RedisEnabled {
		manager.redis = redis.NewClient(&redis.Options{
			Addr:            cfg.RedisAddress,
			Password:        cfg.RedisPassword,
			DB:              cfg.RedisDB,
			DialTimeout:     cfg.RedisDialTimeout,
			ReadTimeout:     cfg.RedisReadTimeout,
			WriteTimeout:    cfg.RedisWriteTimeout,
			MaxRetries:      cfg.RedisMaxRetries,
			MinRetryBackoff: cfg.RedisMinRetryBackoff,
			MaxRetryBackoff: cfg.RedisMaxRetryBackoff,
		})

		if err := manager.redis.Ping(context.Background()).Err(); err != nil {
			// 这里容忍 Redis 不可用的情况，降级为单机缓存
			logger.Warn("Redis 无法连接，降级为仅 BigCache", zap.Error(err))
			_ = manager.redis.Close()
			manager.redis = nil
			manager.redisEnabled = false
			manager.markRedisAlive(false)
		} else {
			manager.markRedisAlive(true)
		}
	} else {
		manager.markRedisAlive(false)
	}

	return manager, nil
}

// Get 会先尝试从 BigCache 命中，再回源 Redis
func (m *Manager) Get(ctx context.Context, key string) (string, HitLevel) {
	if data, err := m.local.Get(key); err == nil {
		m.hitsLocal.Add(1)
		return string(data), HitLocal
	} else if !errors.Is(err, bigcache.ErrEntryNotFound) {
		m.logger.Warn("BigCache 读取失败", zap.Error(err))
	}

	if !m.redisEnabled || m.redis == nil {
		m.misses.Add(1)
		return "", HitNone
	}

	value, err := m.redis.Get(ctx, key).Result()
	switch {
	case err == nil:
		m.hitsRedis.Add(1)
		m.markRedisAlive(true)
		if setErr := m.local.Set(key, []byte(value)); setErr != nil {
			m.logger.Warn("Redis 回填 BigCache 失败", zap.Error(setErr))
		}
		return value, HitRedis
	case errors.Is(err, redis.Nil):
		m.misses.Add(1)
		return "", HitNone
	default:
		m.logger.Warn("Redis 读取失败", zap.Error(err))
		m.markRedisAlive(false)
		m.misses.Add(1)
		return "", HitNone
	}
}

// Set 将数据写入 BigCache，并异步写 Redis
func (m *Manager) Set(ctx context.Context, key string, value string) {
	if err := m.local.Set(key, []byte(value)); err != nil {
		m.logger.Warn("BigCache 写入失败", zap.Error(err))
	}

	if !m.redisEnabled || m.redis == nil {
		return
	}

	go func() {
		childCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := m.redis.Set(childCtx, key, value, m.redisTTL).Err(); err != nil {
			m.logger.Warn("Redis 写入失败", zap.Error(err))
			m.markRedisAlive(false)
			return
		}
		m.markRedisAlive(true)
	}()
}

// Close 主动释放底层资源，便于优雅停机
func (m *Manager) Close() error {
	if m.local != nil {
		if err := m.local.Close(); err != nil {
			m.logger.Warn("BigCache 关闭失败", zap.Error(err))
		}
	}

	if m.redisEnabled && m.redis != nil {
		if err := m.redis.Close(); err != nil {
			m.logger.Warn("Redis 关闭失败", zap.Error(err))
			return err
		}
	}

	return nil
}
