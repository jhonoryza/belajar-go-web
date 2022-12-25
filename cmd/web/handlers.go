package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.labkita.my.id/internal/models"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (app *application) home(resp http.ResponseWriter, req *http.Request) {
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
	params := httprouter.ParamsFromContext(req.Context())
	//id, err := strconv.Atoi(req.URL.Query().Get("id"))
	id, err := strconv.Atoi(params.ByName("id"))
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

func (app *application) snippetCreateForm(resp http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(resp, http.StatusOK, "create.tmpl", data)
}

type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

func (app *application) snippetCreate(resp http.ResponseWriter, req *http.Request) {
	// cek bad request
	err := req.ParseForm()
	if err != nil {
		app.clientError(resp, http.StatusBadRequest)
		return
	}

	// grab body param
	expires, err := strconv.Atoi(req.PostForm.Get("expires"))
	if err != nil {
		app.clientError(resp, http.StatusBadRequest)
		return
	}

	form := snippetCreateForm{
		Title:       req.PostForm.Get("title"),
		Content:     req.PostForm.Get("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}

	/** validation snippets: https://www.alexedwards.net/blog/validation-snippets-for-go
	 * do validation
	 */
	if strings.TrimSpace(form.Title) == "" {
		form.FieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErrors["title"] = "This field cannot be more than 100 characters long"
	}
	if strings.TrimSpace(form.Content) == "" {
		form.FieldErrors["content"] = "This field cannot be blank"
	}
	if form.Expires != 1 && form.Expires != 7 && form.Expires != 365 {
		form.FieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}
	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(resp, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// create data
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(resp, err)
		return
	}

	// redirection
	http.Redirect(resp, req, fmt.Sprintf("/snippets/view/%d", id), http.StatusSeeOther)
}
