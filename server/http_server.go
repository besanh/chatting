package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besanh/chatting/config"
	"github.com/besanh/chatting/middleware/metric"
	"github.com/besanh/chatting/service"
	log "github.com/besanh/logger/logging/slog"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpServer struct {
	Server *gin.Engine
	Port   string
}

func NewServer(envMode, port string) *HttpServer {
	switch envMode {
	case "test":
		gin.SetMode(gin.TestMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": service.API_SERVICE_NAME,
			"version": service.API_VERSION,
			"time":    time.Now().Unix(),
		})
	})

	if envMode != gin.ReleaseMode {
		engine.Use(metric.PrometheusMiddleware())
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	return &HttpServer{
		Server: engine,
		Port:   port,
	}
}

func (s *HttpServer) Start(config config.Config) error {
	addr := fmt.Sprintf(":%s", s.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.Server,
		ReadTimeout:  time.Duration(config.Server.GracefulShutdown.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Server.GracefulShutdown.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.Server.GracefulShutdown.IdleTimeout) * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		var err error
		if config.Server.Mode == "release" {
			err = srv.ListenAndServeTLS(config.Api.TlsCertPath, config.Api.TlsKeyPath)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start HTTP server on %s: %v", addr, err))
		}
	}()

	log.Infof("server runs on %s", addr)

	<-quit
	log.Warn("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Server.GracefulShutdown.WriteTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("failed to shutdown server: %v", err)
		return err
	}

	log.Info("server exited clearly")

	return nil
}
