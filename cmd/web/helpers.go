package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *application) serverError(resp http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Print(trace)

	http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(resp http.ResponseWriter, status int) {
	http.Error(resp, http.StatusText(status), status)
}

func (app *application) notFound(resp http.ResponseWriter) {
	app.clientError(resp, http.StatusNotFound)
}

func (app *application) render(resp http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(resp, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(resp, err)
		return
	}

	resp.WriteHeader(status)

	buf.WriteTo(resp)
}

func (app *application) newTemplateData(req *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(req.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(req),
		CSRFToken:       nosurf.Token(req),
	}
}

func (app *application) decodePostForm(req *http.Request, dst any) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, req.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}
