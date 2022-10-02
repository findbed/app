package router

import (
	"github.com/findbed/app/web"
	"github.com/gin-gonic/gin"
)

func Web(engine *gin.Engine) {
	engine.LoadHTMLGlob("web/tmpls/**/*")
	engine.GET("/", web.RootHandler)
}
