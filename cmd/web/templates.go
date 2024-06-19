package main

import (
	"SnippetAppBook/internal/models"
	"html/template"
	"path/filepath"
	"time"
)

type templateData struct {
	Year      int
	Snippet   *models.Snippet
	Snippets  []*models.Snippet
	Form      any
	Flash     string
	IsAuth    bool
	CSRFToken string
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {

	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("internal/ui/html/pages/*.tmpl")

	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("internal/ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("internal/ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
