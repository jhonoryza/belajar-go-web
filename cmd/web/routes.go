package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	// Route defined
	mux := http.NewServeMux()

	// Static asset Route
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Handler Route
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippets", app.snippetList)
	mux.HandleFunc("/snippets/view", app.snippetView)
	mux.HandleFunc("/snippets/create", app.snippetCreate)

	return mux
}