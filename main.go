package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imega/daemon"
	"github.com/imega/daemon/configuring/env"
	httpserver "github.com/imega/daemon/http-server"
	"github.com/imega/daemon/logging/wrapzerolog"
	"github.com/rs/zerolog"
)

const (
	shutdownTimeout = 15 * time.Second

	appName = "app"
)

func main() {
	logger := wrapzerolog.New(zerolog.New(os.Stderr).With().Logger())

	router := gin.New()

	router.GET("/healthcheck", func(c *gin.Context) { c.Status(204) })

	router.GET("/", func(c *gin.Context) {
		c.Status(200)
		c.Writer.Write([]byte("hello findbed"))
	})

	httpSrv := httpserver.New(
		appName,
		httpserver.WithLogger(logger),
		httpserver.WithHandler(router),
	)

	confReader := env.Once(
		httpSrv.WatcherConfigFunc,
	)

	app, err := daemon.New(logger, confReader)
	if err != nil {
		logger.Errorf("failed to create an instance of daemon, %s", err)
		os.Exit(1)
	}

	logger.Infof("%s is started", appName)

	if err := app.Run(shutdownTimeout); err != nil {
		logger.Errorf("failed to run a daemon, %s", err)
	}

	logger.Infof("%s is stopped", appName)
}
