package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/findbed/app/l10n"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/foolin/goview/supports/gorice"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func WebRouter(engine *gin.Engine) {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateWorkingDirectory},
	}

	localization := l10n.New2()
	translationsBox := conf.MustFindBox("translations")
	translationsBox.Walk("/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".ftl") {
			return nil
		}

		fileName := info.Name()
		log.Printf("info: %s", fileName[:len(fileName)-len(filepath.Ext(fileName))])

		file, err := translationsBox.Open(fileName)
		if err != nil {
			return fmt.Errorf("failed to open translate, %w", err)
		}

		dict, err := l10n.LoadDict(file)
		if err != nil {
			return fmt.Errorf("failed to load dictionary, %w", err)
		}

		localization.AddDict(fileName[:len(fileName)-len(filepath.Ext(fileName))], dict)

		return nil
	})

	localization.MakeCatalog()

	printer := message.NewPrinter(language.English)

	basic := gorice.NewWithConfig(
		conf.MustFindBox("web/views"),
		goview.Config{
			Root:      "views",
			Extension: ".html",
			Master:    "layout",
			Funcs: template.FuncMap{
				"current_year": func() string {
					return time.Now().Format("2006")
				},
				"l10n": printer.Sprintf,
			},
			DisableCache: false,
		},
	)
	engine.HTMLRender = ginview.Wrap(basic)

	engine.GET("/", setLocale(printer, localization), RootHandler)

	engine.StaticFS("/assets", conf.MustFindBox("web/assets").HTTPBox())
}

func setLocale(printer *message.Printer, locale *l10n.L10N) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lngHeader := ctx.GetHeader("Accept-Language")
		lngTags, _, _ := language.ParseAcceptLanguage(lngHeader)

		matcher := language.NewMatcher(lngTags)
		lngTag, _, _ := matcher.Match()

		*printer = *message.NewPrinter(lngTag)
	}
}
