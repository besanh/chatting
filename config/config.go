package config

import (
	"time"

	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	"github.com/nats-io/nats.go"
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
