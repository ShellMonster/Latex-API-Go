package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Server 用于描述 HTTP 服务自身的行为配置
type Server struct {
	Address           string        `mapstructure:"address"`
	Prefork           bool          `mapstructure:"prefork"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	RequestTimeout    time.Duration `mapstructure:"request_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	MaxRequestBodyMB  int           `mapstructure:"max_request_body_mb"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

// Log 用于描述日志文件的滚动策略
type Log struct {
	Filename   string `mapstructure:"filename"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
	Level      string `mapstructure:"level"`
}

// Cache 用于配置一级与二级缓存策略
type Cache struct {
	LocalLifeWindow      time.Duration `mapstructure:"local_life_window"`
	LocalCleanWindow     time.Duration `mapstructure:"local_clean_window"`
	LocalHardMaxCacheMB  int           `mapstructure:"local_hard_max_cache_mb"`
	RedisEnabled         bool          `mapstructure:"redis_enabled"`
	RedisAddress         string        `mapstructure:"redis_address"`
	RedisPassword        string        `mapstructure:"redis_password"`
	RedisDB              int           `mapstructure:"redis_db"`
	RedisDialTimeout     time.Duration `mapstructure:"redis_dial_timeout"`
	RedisReadTimeout     time.Duration `mapstructure:"redis_read_timeout"`
	RedisWriteTimeout    time.Duration `mapstructure:"redis_write_timeout"`
	RedisTTL             time.Duration `mapstructure:"redis_ttl"`
	RedisMaxRetries      int           `mapstructure:"redis_max_retries"`
	RedisMinRetryBackoff time.Duration `mapstructure:"redis_min_retry_backoff"`
	RedisMaxRetryBackoff time.Duration `mapstructure:"redis_max_retry_backoff"`
}

// Config 汇总服务启动所需的所有配置模块
type Config struct {
	Server Server `mapstructure:"server"`
	Log    Log    `mapstructure:"log"`
	Cache  Cache  `mapstructure:"cache"`
}

// Load 负责读取配置文件与环境变量，返回结构化配置
func Load() (Config, error) {
	setDefaults()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("mathsvg")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	if err := ensureLogDir(cfg.Log.Filename); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// setDefaults 为避免新环境缺少配置文件，预置一套默认值
func setDefaults() {
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.prefork", true)
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "5s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.request_timeout", "3s")
	viper.SetDefault("server.shutdown_timeout", "5s")
	viper.SetDefault("server.max_request_body_mb", 5)
	viper.SetDefault("server.enable_compression", true)

	viper.SetDefault("log.filename", "logs/server.log")
	viper.SetDefault("log.max_size_mb", 50)
	viper.SetDefault("log.max_backups", 7)
	viper.SetDefault("log.max_age_days", 30)
	viper.SetDefault("log.compress", true)
	viper.SetDefault("log.level", "info")

	viper.SetDefault("cache.local_life_window", "10m")
	viper.SetDefault("cache.local_clean_window", "1m")
	viper.SetDefault("cache.local_hard_max_cache_mb", 256)
	viper.SetDefault("cache.redis_enabled", false)
	viper.SetDefault("cache.redis_address", "localhost:6379")
	viper.SetDefault("cache.redis_password", "")
	viper.SetDefault("cache.redis_db", 0)
	viper.SetDefault("cache.redis_dial_timeout", "500ms")
	viper.SetDefault("cache.redis_read_timeout", "2s")
	viper.SetDefault("cache.redis_write_timeout", "2s")
	viper.SetDefault("cache.redis_ttl", "168h")
	viper.SetDefault("cache.redis_max_retries", 2)
	viper.SetDefault("cache.redis_min_retry_backoff", "100ms")
	viper.SetDefault("cache.redis_max_retry_backoff", "500ms")
}

// ensureLogDir 在加载配置时提前确保日志目录存在
func ensureLogDir(filename string) error {
	if filename == "" {
		return nil
	}
	dir := filepath.Dir(filename)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
