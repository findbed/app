package web

import (
	"log"
	"net/http"

	"github.com/foolin/goview"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func RootHandler(ctx *gin.Context) {
	err := goview.Render(ctx.Writer, http.StatusOK, "index.tmpl", goview.M{
		"title": "Main website",
		"l10n":  extractL10N(ctx),
	})
	if err != nil {
		log.Printf("======== %s", err.Error())
	}
}

func extractL10N(ctx *gin.Context) interface{} {
	val, isExist := ctx.Get("prt")
	if !isExist {
		return message.NewPrinter(language.English).Sprintf
	}

	printer, ok := val.(*message.Printer)
	if !ok {
		return message.NewPrinter(language.English).Sprintf
	}

	return printer.Sprintf
}
