package main

import (
	"bytes"
	_ "embed"
	"text/template"

	"gopkg.in/yaml.v3"
)

type predefinedTexts struct {
	texts map[string]string
	cache map[string]*template.Template
}

func (t *predefinedTexts) Make(name string, opts ...any) string {

	var data any
	switch len(opts) {
	case 1:
		data = opts[0]
	case 0:
	default:
		data = opts
	}

	tplData := t.texts[name]

	var tp *template.Template

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"escape": func(s string) (string, error) {
			res := []rune{}

			for _, r := range s {
				switch r {
				case 13, 10:
					r = '↵'
				}
				res = append(res, r)
			}
			return string(res), nil
		},
		// The name "title" is what the function will be called in the template text.
		"lescape": func(s string) (string, error) {
			res := []rune{}

			for _, r := range s {
				switch r {
				case 13, 10:
					r = ' '
				}
				res = append(res, r)
			}
			return string(res), nil
		},
		"version": func() (string, error) {
			return Version, nil
		},
	}

	if tpl, ok := t.cache[name]; ok {
		tp = tpl
	} else {
		tp = template.Must(template.New(name).Funcs(funcMap).Parse(tplData))
		t.cache[name] = tp
	}

	output := bytes.NewBufferString("")
	if err := tp.Execute(output, data); err != nil {
		panic(err)
	}
	return output.String()
}

//go:embed predefined.yaml
var predefinedTextsData []byte

func initTexts() *predefinedTexts {
	texts := predefinedTexts{
		cache: make(map[string]*template.Template),
		texts: make(map[string]string),
	}
	yaml.Unmarshal(predefinedTextsData, &texts.texts)

	return &texts
}
