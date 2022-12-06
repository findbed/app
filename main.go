package main

import (
	"os"
	"time"

	"github.com/findbed/app/api"
	"github.com/findbed/app/web"
	"github.com/gin-gonic/gin"
	"github.com/imega/daemon"
	"github.com/imega/daemon/configuring/env"
	httpserver "github.com/imega/daemon/http-server"
	"github.com/imega/daemon/logging/wrapzerolog"
	"github.com/rs/zerolog"
)

//+start_at:<=6 +end_at:>=10
// TAKE BOOKING

// select * from timeslot where start_at<=10 and end_at>=11
// update timeslot set end_at=10 where id=?
// insert timeslot (lot_id,start_at,end_at) values (?, 11, 65536)

// RETURN BOOKING

// Return booking with center-append
// select * from timeslot where end_at=20 or start_at=21
// if return empty you should insert record with start_at=20 and end_at=21

// if return booking timeslot left-append
// select * from timeslot where end_at=19 or start_at=20
// +----+--------+----------+--------+
// | id | lot_id | start_at | end_at |
// +----+--------+----------+--------+
// |  2 |      1 |       15 |     19 |
// +----+--------+----------+--------+
// update timeslot set end_at=20 where id=2

// return booking with right-append
// select * from timeslot where end_at=21 or start_at=22
// +----+--------+----------+--------+
// | id | lot_id | start_at | end_at |
// +----+--------+----------+--------+
// |  5 |      1 |       22 |  65535 |
// +----+--------+----------+--------

const (
	shutdownTimeout = 15 * time.Second

	appName = "app"
)

func main() {
	logger := wrapzerolog.New(zerolog.New(os.Stderr).With().Logger())

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.GET("/healthcheck", func(c *gin.Context) { c.Status(204) })

	web.WebRouter(engine)
	api.APIRouter(engine)

	httpSrv := httpserver.New(
		appName,
		httpserver.WithLogger(logger),
		httpserver.WithHandler(engine),
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
