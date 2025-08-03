package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/KMankowski/snippetbox/internal/models"
)

type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	fieldErrors map[string]string
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.db.Latest()
	if err != nil {
		app.handleServerError(w, r, err)
		return
	}

	// for _, snippet := range snippets {
	// 	fmt.Fprintf(w, "%+v\n", snippet)
	// }

	data := newTemplateData()
	data.Snippets = snippets

	app.renderPage(w, r, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := strconv.Atoi(rawId)
	if err != nil || id < 0 {
		http.NotFound(w, r)
		return
	}

	snippet, err := app.db.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.handleServerError(w, r, err)
		}
		return
	}

	data := newTemplateData()
	data.Snippet = snippet

	app.renderPage(w, r, http.StatusOK, "view.tmpl", data)

	// html/template package only allows passing in ONE piece of dynamic data! crazy.
	// err = template.ExecuteTemplate(w, "base", snippet)
	// if err != nil {
	// 	app.handleServerError(w, r, err)
	// 	return
	// }

	// w.Header().Set("content-type", "application/json")
	// fmt.Fprintf(w, "Display a snippet with id %v...", id)
	// w.Write([]byte(fmt.Sprintf("Display a snippet with id %v...", r.PathValue("id"))))
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	app.renderPage(w, r, http.StatusOK, "create.tmpl", newTemplateData())
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// title := "0 snail"
	// content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	// expires := 7

	// id, err := app.db.Insert(title, content, expires)
	// if err != nil {
	// app.handleServerError(w, r, err)
	// return
	// }

	// http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)

	// r.Body = http.MaxBytesReader(w, r.Body, 10)
	err := r.ParseForm()
	if err != nil {
		app.handleClientError(w, http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.handleClientError(w, http.StatusBadRequest)
		return
	}

	fieldErrors := make(map[string]string)

	if strings.TrimSpace(title) == "" {
		fieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(title) > 100 {
		fieldErrors["title"] = "This field cannot contain more than 100 characters"
	}

	if strings.TrimSpace(content) == "" {
		fieldErrors["content"] = "This field cannot be blank"
	}

	if expires != 1 && expires != 7 && expires != 365 {
		fieldErrors["expires"] = "This field must equal 1, 7, or 365"
	}

	if len(fieldErrors) > 0 {
		templateData := newTemplateData()
		templateData.Form = snippetCreateForm{
			title,
			content,
			expires,
			fieldErrors,
		}
		app.renderPage(w, r, http.StatusBadRequest, "create.tmpl", templateData)
		fmt.Fprint(w, fieldErrors)
		return
	}

	id, err := app.db.Insert(title, content, expires)
	if err != nil {
		app.handleServerError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
