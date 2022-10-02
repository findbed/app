package main

import (
	"net/http"
	"os"
	"time"

	"github.com/imega/daemon"
	"github.com/imega/daemon/configuring/env"
	healthhttp "github.com/imega/daemon/health/http"
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

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/healthcheck",
		healthhttp.HandlerFunc(
			healthhttp.WithHealthCheckFuncs(
				func() bool { return true },
			),
		),
	)

	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("findbed.ru"))
		},
	)

	httpSrv := httpserver.New(
		appName,
		httpserver.WithLogger(logger),
		httpserver.WithHandler(mux),
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
