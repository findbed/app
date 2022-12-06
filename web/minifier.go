package web

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foolin/goview"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func minifyTemplate() goview.FileHandler {
	minifier := minify.New()
	minifier.AddFunc("text/html", html.Minify)

	return func(config goview.Config, tplFile string) (string, error) {
		filename := tplFile + config.Extension
		path := filepath.Join(config.Root, filename)

		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s, %s", filename, err)
		}

		data, err = minifier.Bytes("text/html", data)
		if err != nil {
			return "",
				fmt.Errorf("failed to minify template %s, %s", filename, err)
		}

		return string(data), nil
	}
}
