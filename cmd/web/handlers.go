package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox.labkita.my.id/internal/models"
	"strconv"
)

func (app *application) home(resp http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(resp, req)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(resp, err)
		return
	}
	data := app.newTemplateData(req)
	data.Snippets = snippets
	app.render(resp, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetList(resp http.ResponseWriter, req *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(resp, err)
		return
	}
	for _, snippet := range snippets {
		fmt.Fprintf(resp, "%+v\n", snippet)
	}
}

func (app *application) snippetView(resp http.ResponseWriter, req *http.Request) {
	//validation id
	id, err := strconv.Atoi(req.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(resp)
		return
	}

	//query by id
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(resp)
		} else {
			app.serverError(resp, err)
		}
		return
	}
	data := app.newTemplateData(req)
	data.Snippet = snippet

	app.render(resp, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.Header().Set("Allow", http.MethodPost)
		app.clientError(resp, http.StatusMethodNotAllowed)
		return
	}

	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(resp, err)
		return
	}

	http.Redirect(resp, req, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
