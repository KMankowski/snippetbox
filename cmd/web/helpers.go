package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func (app *application) handleServerError(w http.ResponseWriter, r *http.Request, err error) {
	requestMethod := r.Method
	requestUri := r.URL.RequestURI()
	stackTrace := string(debug.Stack())

	app.logger.Error(err.Error(),
		slog.String("method", requestMethod),
		slog.String("uri", requestUri),
		slog.String("callstack", stackTrace))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) handleClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) renderPage(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	template, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.handleServerError(w, r, err)
		return
	}

	htmlBuffer := new(bytes.Buffer)

	err := template.ExecuteTemplate(htmlBuffer, "base", data)
	if err != nil {
		app.handleServerError(w, r, err)
		return
	}

	w.WriteHeader(status)
	htmlBuffer.WriteTo(w)
}
