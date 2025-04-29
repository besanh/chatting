package main

import (
	"errors"
	_ "net/http/pprof"
	"slices"

	"github.com/besanh/chatbot_gpt/config"
	server "github.com/besanh/chatbot_gpt/server"
	"github.com/besanh/chatbot_gpt/service"
	"github.com/gin-gonic/gin"
)

var cfg config.Config

func init() {
	config.InitConfig(&cfg)
}

func main() {
	if !slices.Contains([]string{gin.ReleaseMode, gin.DebugMode, gin.TestMode}, cfg.Server.Mode) {
		panic(errors.New("env was incorrect"))
	}

	// Init server http
	httpServer := server.NewServer(cfg.Server.Mode, cfg.Server.Port)
	initLayers(httpServer.Server)
}

func initLayers(httpRouter *gin.Engine) {
	// Global config
	service.API_SERVICE_NAME = cfg.Api.ApiServiceName
	service.API_VERSION = cfg.Api.ApiVersion
}
