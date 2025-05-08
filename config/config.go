package config

import (
	"encoding/json"
	"strings"
	"time"

	cache "github.com/besanh/chatting/common/caching"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	messagequeue "github.com/besanh/chatting/pkg/message_queue"
	"github.com/besanh/chatting/pkg/mongodb"
	"github.com/besanh/chatting/pkg/redis"
	"github.com/besanh/chatting/pkg/sqlclient"
	"github.com/besanh/chatting/repository"
	log "github.com/besanh/logger/logging/slog"
	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type (
	Config struct {
		ConfigDir string `envDefault:"./config/config.yml"`
		Server    Server `mapstructure:"server"`
		Api       Api    `mapstructure:"api"`
		Pkg       Pkg    `mapstructure:"pkg"`
	}
	Server struct {
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
		RateLimit      int    `mapstructure:"rate_limit"`
	}

	Pkg struct {
		Openai struct {
			Enable bool     `mapstructure:"enable"`
			ApiKey string   `mapstructure:"api_key"`
			Models []string `mapstructure:"models"`
		} `mapstructure:"openai"`

		Translate struct {
			Enable  bool     `mapstructure:"enable"`
			Url     []string `mapstructure:"url"`
			Timeout int      `mapstructure:"timeout"`
			Retry   int      `mapstructure:"retry"`
		} `mapstructure:"translate"`

		Redis struct {
			Enable bool   `mapstructure:"enable"`
			Dsn    string `mapstructure:"dsn"`
		} `mapstructure:"redis"`

		PostgreSql struct {
			Enable       bool   `mapstructure:"enable"`
			Username     string `mapstructure:"username"`
			Password     string `mapstructure:"password"`
			Host         string `mapstructure:"host"`
			Port         int    `mapstructure:"port"`
			Database     string `mapstructure:"database"`
			DialTimeout  int    `mapstructure:"dial_timeout"`
			ReadTimeout  int    `mapstructure:"read_timeout"`
			WriteTimeout int    `mapstructure:"write_timeout"`
			Timeout      int    `mapstructure:"timeout"`
			PoolSize     int    `mapstructure:"pool_size"`
			MaxOpenConns int    `mapstructure:"max_open_conns"`
			MaxIdleConns int    `mapstructure:"max_idle_conns"`
		} `mapstructure:"postgresql"`

		MongoDb struct {
			Enable        bool   `mapstructure:"enable"`
			Username      string `mapstructure:"username"`
			Password      string `mapstructure:"password"`
			Host          string `mapstructure:"host"`
			Port          int    `mapstructure:"port"`
			Database      string `mapstructure:"database"`
			DefaultAuthDb string `mapstructure:"default_auth_db"`
		} `mapstructure:"mongo_db"`

		NatJetstream struct {
			Enable         bool          `mapstructure:"enable"`
			Dsn            string        `mapstructure:"dsn"`
			PublishTimeout time.Duration `mapstructure:"publish_timeout"`
			Js             *nats.JetStreamContext
			CB             *circuitbreaker.CBSetting `mapstructure:"cb"`
		} `mapstructure:"nat_jetstream"`

		Oauth2 Oauth2 `mapstructure:"oauth2"`
	}

	Oauth2 struct {
		Google GoogleConfig `mapstructure:"google"`
	}

	GoogleConfigYAML struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scope        []string `mapstructure:"scope"`
		Endpoint     struct {
			AuthURL  string `mapstructure:"auth_url"`
			TokenURL string `mapstructure:"token_url"`
		} `mapstructure:"endpoint"`
		RedirectUrl string `mapstructure:"redirect_url"`
		UserInfoUrl string `mapstructure:"user_info_url"`
		RevokeUrl   string `mapstructure:"revoke_url"`
	}

	GoogleConfig struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scope        []string `mapstructure:"scope"`
		Endpoint     oauth2.Endpoint
		RedirectUrl  string `mapstructure:"redirect_url"`
		UserInfoUrl  string `mapstructure:"user_info_url"`
		RevokeUrl    string `mapstructure:"revoke_url"`
	}
)

func InitConfig(cfg *Config, mongoDB mongodb.IMongoDBClient) {
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

	loadGoogleOAuthConfig(cfg)
	loadModelAI(cfg)

	initLogger(cfg)
	if cfg.Pkg.Redis.Enable {
		initRedis(cfg)
	}

	if cfg.Pkg.MongoDb.Enable {
		initMongoDb(cfg, mongoDB)
	}

	if cfg.Pkg.NatJetstream.Enable {
		initNatsJetstream(cfg)
	}

	if cfg.Pkg.PostgreSql.Enable {
		initSql(cfg)
	}

	registerMetrics()
}

func loadGoogleOAuthConfig(cfg *Config) {
	var yamlCfg GoogleConfigYAML
	if err := viper.UnmarshalKey("pkg.oauth2.google", &yamlCfg); err != nil {
		log.Fatal("unmarshal failed: %v", err)
	}

	cfg.Pkg.Oauth2.Google = GoogleConfig{
		ClientId:     yamlCfg.ClientId,
		ClientSecret: yamlCfg.ClientSecret,
		Scope:        yamlCfg.Scope,
		RedirectUrl:  yamlCfg.RedirectUrl,
		UserInfoUrl:  yamlCfg.UserInfoUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  yamlCfg.Endpoint.AuthURL,
			TokenURL: yamlCfg.Endpoint.TokenURL,
		},
		RevokeUrl: yamlCfg.RevokeUrl,
	}
}

func loadModelAI(cfg *Config) {
	if raw := viper.GetString("pkg.openai.models"); raw != "" {
		var models []string
		if err := json.Unmarshal([]byte(raw), &models); err != nil {
			log.Errorf("invalid JSON in PKG_OPENAI_MODELS: %v", err)
			panic(err)
		} else {
			cfg.Pkg.Openai.Models = models
		}
	}
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

func initLogger(cfg *Config) {
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

func initRedis(cfg *Config) {
	redisClient, err := redis.NewRedis(redis.RedisConfig{Dsn: cfg.Pkg.Redis.Dsn})
	if err != nil {
		panic(err)
	}

	cache.RCache = cache.NewRedisCache(redisClient.GetClient())
}

func initMongoDb(cfg *Config, mongoDB mongodb.IMongoDBClient) {
	mongodbConfig := mongodb.MongoDBConfig{
		Username:      cfg.Pkg.MongoDb.Username,
		Password:      cfg.Pkg.MongoDb.Password,
		Host:          cfg.Pkg.MongoDb.Host,
		Port:          cfg.Pkg.MongoDb.Port,
		Database:      cfg.Pkg.MongoDb.Database,
		DefaultAuthDb: cfg.Pkg.MongoDb.DefaultAuthDb,
	}

	var err error
	var db mongodb.IMongoDBClient
	db, err = mongodb.NewMongoDBClient(mongodbConfig)
	if err != nil {
		// If the connection string is invalid, panic
		log.Errorf("mongodb connect error: %v", err)
		panic(err)
	}

	mongoDB = db
}

func initNatsJetstream(cfg *Config) {
	nc, err := nats.Connect(cfg.Pkg.NatJetstream.Dsn)
	if err != nil {
		log.Fatalf("nats connect failed: %v", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("nc jetStream failed: %v", err)
	}

	nat := &messagequeue.NatsJetStream{
		Cfg: messagequeue.Config{
			Host:           cfg.Pkg.NatJetstream.Dsn,
			PublishTimeout: cfg.Pkg.NatJetstream.PublishTimeout,
		},
		Nc: nc,
		Js: js,
	}
	cfg.Pkg.NatJetstream.Js = &nat.Js

	// Connect to NATS JetStream
	if err := nat.Connect(); err != nil {
		// If the connection string is invalid, panic
		log.Errorf("nats jetstream connect error: %v", err)
		panic(err)
	}

	// Init events
	initEvents(cfg)
}

func initSql(cfg *Config) {
	sqlClientConfig := sqlclient.SqlConfig{
		Host:         cfg.Pkg.PostgreSql.Host,
		Database:     cfg.Pkg.PostgreSql.Database,
		Username:     cfg.Pkg.PostgreSql.Username,
		Password:     cfg.Pkg.PostgreSql.Password,
		Port:         cfg.Pkg.PostgreSql.Port,
		DialTimeout:  cfg.Pkg.PostgreSql.DialTimeout,
		ReadTimeout:  cfg.Pkg.PostgreSql.ReadTimeout,
		WriteTimeout: cfg.Pkg.PostgreSql.WriteTimeout,
		Timeout:      cfg.Pkg.PostgreSql.Timeout,
		PoolSize:     cfg.Pkg.PostgreSql.PoolSize,
		MaxOpenConns: cfg.Pkg.PostgreSql.MaxOpenConns,
		MaxIdleConns: cfg.Pkg.PostgreSql.MaxIdleConns,
		Driver:       sqlclient.POSTGRESQL,
	}
	repository.DBConn = sqlclient.NewSqlClient(sqlClientConfig)
}

func initEvents(cfg *Config) {
	streamName := "USER_EVENT"
	if _, err := (*cfg.Pkg.NatJetstream.Js).StreamInfo(streamName); err == nil {
		log.Info("stream already exists")
		return
	}

	subject := []string{"user_events.*"}
	_, err := (*cfg.Pkg.NatJetstream.Js).AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  subject,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("stream created")
	return
}
