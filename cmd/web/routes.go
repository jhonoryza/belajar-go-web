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

	// route middleware
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Handler Route
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippets", dynamic.ThenFunc(app.snippetList))
	router.Handler(http.MethodGet, "/snippets/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/snippets/create", dynamic.ThenFunc(app.snippetCreateForm))
	router.Handler(http.MethodPost, "/snippets/create", dynamic.ThenFunc(app.snippetCreate))

	// global middleware
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
