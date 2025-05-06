package main

import (
	"context"
	"errors"
	_ "net/http/pprof"
	"slices"
	"time"

	v1 "github.com/besanh/chatting/api/v1"
	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/pkg/mongodb"
	pkgOauth2 "github.com/besanh/chatting/pkg/oauth2"
	"github.com/besanh/chatting/repository"
	"github.com/besanh/chatting/server"
	"github.com/besanh/chatting/service"
	"github.com/gin-gonic/gin"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Repository
	repository.InitTables(ctx, repository.DBConn)
	repository.InitRepositories()

	// Api
	oauth2Client := pkgOauth2.NewOAuth2(cfg.Pkg.Oauth2)
	v1.NewUsers(httpRouter, service.NewUser(repository.UserRepo, oauth2Client, cfg))

	// Service
	service.NewUser(repository.UserRepo, oauth2Client, cfg)
}
