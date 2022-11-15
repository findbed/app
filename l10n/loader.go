// Copyright © 2022 Dmitry Stoletov <info@imega.ru>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package l10n

import (
	"fmt"
	"io"
	"strings"

	"github.com/findbed/app/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/ini.v1"
)

type L10N struct {
	dictionaries map[string]catalog.Dictionary
}

func New2() *L10N {
	return &L10N{
		dictionaries: map[string]catalog.Dictionary{},
	}
}

func (l10n *L10N) AddDict(lang string, cat catalog.Dictionary) {
	l10n.dictionaries[lang] = cat
}

func (l10n *L10N) MakeCatalog() {
	cat, err := catalog.NewFromMap(
		l10n.dictionaries,
		catalog.Fallback(language.English),
	)
	if err != nil {
		return
	}

	l10n.dictionaries = nil
	message.DefaultCatalog = cat
}

func (l10n *L10N) GetPrinter(lang string) *message.Printer {
	lngTag, err := language.Parse(lang)
	if err != nil {
		return message.NewPrinter(language.English)
	}

	return message.NewPrinter(lngTag)
}

func New(lang string, r io.ReadCloser) (*message.Printer, error) {
	lngTag, err := language.Parse(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a lang code, %w", err)
	}

	if r == nil {
		r = io.NopCloser(strings.NewReader(""))
	}

	cat, err := loadCatalog(lngTag, r)
	if err != nil {
		return nil, fmt.Errorf("failed to load a lang-catalog, %w", err)
	}

	message.DefaultCatalog = cat
	printer := message.NewPrinter(lngTag)

	return printer, nil
}

func loadCatalog(lang language.Tag, rd io.ReadCloser) (catalog.Catalog, error) {
	fallbackDict, err := LoadDict(strings.NewReader(translations.Fallback()))
	if err != nil {
		return nil, fmt.Errorf("failed to load fallback dictionary, %w", err)
	}

	dict := map[string]catalog.Dictionary{
		language.English.String(): fallbackDict,
	}

	currentDict, err := LoadDict(rd)
	if err != nil {
		return nil, fmt.Errorf("failed to load current dictionary, %w", err)
	}

	dict[lang.String()] = currentDict

	cat, err := catalog.NewFromMap(dict, catalog.Fallback(language.English))
	if err != nil {
		return nil, fmt.Errorf("failed to create catalog, %w", err)
	}

	return cat, nil
}

func LoadDict(src interface{}) (*Dictionary, error) {
	file, err := ini.Load(src)
	if err != nil {
		return nil, fmt.Errorf("failed to load dictionary, %w", err)
	}

	mapDict := file.Section("").KeysHash()

	return &Dictionary{Src: mapDict}, nil
}
