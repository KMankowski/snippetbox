package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/KMankowski/snippetbox/internal/models"
)

// This contains all of the dynamic data that we want to pass into our templates
type templateData struct {
	CurrentYear int
	Snippet     models.Snippet
	Snippets    []models.Snippet
	Form        snippetCreateForm
}

func getReadableDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

func newTemplateData() templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	templateFunctions := template.FuncMap{
		"getReadableDate": getReadableDate,
	}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		t, err := template.New(name).Funcs(templateFunctions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		t, err = t.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		t, err = t.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = t
	}

	return cache, nil
}
