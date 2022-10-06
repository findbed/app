package web

import (
	"html/template"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/foolin/goview/supports/gorice"
	"github.com/gin-gonic/gin"
)

func WebRouter(engine *gin.Engine) {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateWorkingDirectory},
	}

	basic := gorice.NewWithConfig(
		conf.MustFindBox("views"),
		goview.Config{
			Root:      "views",
			Extension: ".html",
			Master:    "layout",
			Funcs: template.FuncMap{
				"current_year": func() string {
					return time.Now().Format("2006")
				},
			},
			DisableCache: false,
		},
	)
	engine.HTMLRender = ginview.Wrap(basic)

	engine.GET("/", RootHandler)

	engine.StaticFS("/assets", conf.MustFindBox("assets").HTTPBox())
}
