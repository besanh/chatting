package main

import (
	"errors"
	_ "net/http/pprof"
	"slices"

	v1 "github.com/besanh/chatting/api/v1"
	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/external/translation"
	circuitbreaker "github.com/besanh/chatting/pkg/circuit_breaker"
	"github.com/besanh/chatting/pkg/mongodb"
	pkgOauth2 "github.com/besanh/chatting/pkg/oauth2"
	"github.com/besanh/chatting/repository"
	"github.com/besanh/chatting/server"
	"github.com/besanh/chatting/service"
	log "github.com/besanh/logger/logging/slog"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker/v2"
)

var (
	cfg     config.Config
	mongoDB mongodb.IMongoDBClient
)

func init() {
	config.InitConfig(&cfg, mongoDB)
}

func main() {
	if !slices.Contains([]string{gin.ReleaseMode, gin.DebugMode, gin.TestMode}, cfg.Server.Mode) {
		panic(errors.New("env was incorrect"))
	}

	// Init server http
	httpServer := server.NewServer(cfg.Server.Mode, cfg.Server.Port)
	initLayers(httpServer.Server)

	httpServer.Start(cfg)
}

func initLayers(httpRouter *gin.Engine) {
	// Global config
	service.API_SERVICE_NAME = cfg.Api.ApiServiceName
	service.API_VERSION = cfg.Api.ApiVersion

	// Repository
	repository.InitRepositories()

	// Service
	userService := service.NewUser(repository.UserRepo, pkgOauth2.NewOAuth2(cfg.Pkg.Oauth2), cfg)
	chatService := service.NewChat(repository.ChatRepo, repository.ChatMemberRepo)

	// Api
	v1.NewUsers(httpRouter, userService)
	v1.NewChat(httpRouter, chatService)

	cbSetting := circuitbreaker.CBSetting{
		CBName:     "SubscribeServiceCB",
		MaxRequest: 5,
		Interval:   cfg.Pkg.NatJetstream.CB.Interval,
		TimeOut:    cfg.Pkg.NatJetstream.CB.TimeOut,
		MaxTripCB:  3,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Infof("circuit %s: %s -> %s", name, from.String(), to.String())
		},
		IsSuccessful: func(err error) bool { return err == nil },
	}
	v1.NewWs(httpRouter, cfg, service.NewSubscriberService(cfg, &cbSetting, translation.NewGoogleTranslate(cfg)))
}
