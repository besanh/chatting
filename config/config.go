package config

import (
	"encoding/json"
	"strings"

	"github.com/besanh/chatbot_gpt/common/caching"
	"github.com/besanh/chatbot_gpt/pkg/redis"
	log "github.com/besanh/logger/logging/slog"
	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/spf13/viper"
)

type Config struct {
	ConfigDir string `envDefault:"./config/config.yml"`
	Server    struct {
		Port             string `mapstructure:"port"`
		Mode             string `mapstructure:"mode"`
		LogLevel         string `mapstructure:"log_level"`
		LogFile          string `mapstructure:"log_file"`
		GracefulShutdown struct {
			ShutdownTime int64 `mapstructure:"shutdown_time"`
			ReadTimeout  int64 `mapstructure:"read_timeout"`
			WriteTimeout int64 `mapstructure:"write_timeout"`
			IdleTimeout  int64 `mapstructure:"idle_timeout"`
		} `mapstructure:"graceful_shutdown"`
	}

	Api struct {
		ApiServiceName string `mapstructure:"api_service_name"`
		ApiVersion     string `mapstructure:"api_version"`
		TlsCertPath    string `mapstructure:"tls_cert_path"`
		TlsKeyPath     string `mapstructure:"tls_key_path"`
	} `mapstructure:"api"`

	Pkg struct {
		Openai struct {
			ApiKey string   `mapstructure:"api_key"`
			Models []string `mapstructure:"models"`
		} `mapstructure:"openai"`

		Redis struct {
			Dsn string `mapstructure:"dsn"`
		} `mapstructure:"redis"`
	}
}

func InitConfig(cfg *Config) {
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	viper.SetConfigFile(cfg.ConfigDir)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	if raw := viper.GetString("pkg.openai.models"); raw != "" {
		var models []string
		if err := json.Unmarshal([]byte(raw), &models); err != nil {
			log.Errorf("invalid JSON in PKG_OPENAI_MODELS: %v", err)
			panic(err)
		} else {
			cfg.Pkg.Openai.Models = models
		}
	}

	initLogger(*cfg)
	initRedis(*cfg)

	registerMetrics()
}

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func registerMetrics() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		HttpRequestsTotal,
		HttpRequestDuration,
		collectors.NewGoCollector(), // Go runtime metrics
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), // Process metrics
	)
}

func initLogger(cfg Config) {
	logLevel := log.LEVEL_DEBUG
	switch cfg.Server.LogLevel {
	case "debug":
		logLevel = log.LEVEL_DEBUG
	case "info":
		logLevel = log.LEVEL_INFO
	case "error":
		logLevel = log.LEVEL_ERROR
	case "warn":
		logLevel = log.LEVEL_WARN
	}
	opts := []log.Option{}
	opts = append(opts, log.WithLevel(logLevel),
		log.WithRotateFile(cfg.Server.LogFile),
		log.WithFileSource(),
	)

	log.SetLogger(log.NewSLogger(opts...))
}

func initRedis(cfg Config) {
	redisClient, err := redis.NewRedis(redis.RedisConfig{Dsn: cfg.Pkg.Redis.Dsn})
	if err != nil {
		panic(err)
	}

	caching.RCache = caching.NewRedisCache(redisClient.GetClient())
}
