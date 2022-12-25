package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	// Route defined
	router := httprouter.New()

	// Static asset Route
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Custom Handler
	router.NotFound = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		app.notFound(resp)
	})

	// Handler Route
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/snippets", app.snippetList)
	router.HandlerFunc(http.MethodGet, "/snippets/view/:id", app.snippetView)
	router.HandlerFunc(http.MethodGet, "/snippets/create", app.snippetCreateForm)
	router.HandlerFunc(http.MethodPost, "/snippets/create", app.snippetCreate)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
