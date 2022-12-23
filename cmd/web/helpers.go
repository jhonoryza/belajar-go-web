package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
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
