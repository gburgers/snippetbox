package main

import (
	"net/http"

	"github.com/gburgers/snippetbox/ui"
)

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux./ The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// fileServer := http.FileServer(http.Dir("./ui/static/"))
	// mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	mux.Handle("GET /{$}", app.sessionManager.LoadAndSave(http.HandlerFunc(app.home)))
	mux.Handle("GET /snippet/view/{id}", app.sessionManager.LoadAndSave(http.HandlerFunc(app.snippetView)))
	mux.Handle("GET /snippet/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.snippetCreate)))
	mux.Handle("POST /snippet/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.snippetCreatePost)))

	// Wrap the existing chain with the recoverPanic middleware.
	return app.recoverPanic(app.logRequest(commonHeaders(mux)))
}
