package main

import (
	"SnippetAppBook/internal/ui"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *Application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))

	router.Handler(http.MethodGet, "/Static/*filepath", fileServer)
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/view/:id", dynamic.ThenFunc(app.view))
	//router.Handler(http.MethodGet, "/test/1", dynamic.ThenFunc(app.test))

	router.Handler(http.MethodGet, "/signup", dynamic.ThenFunc(app.signUp))
	router.Handler(http.MethodPost, "/signup", dynamic.ThenFunc(app.signUpPost))

	router.Handler(http.MethodGet, "/login", dynamic.ThenFunc(app.login))
	router.Handler(http.MethodPost, "/login", dynamic.ThenFunc(app.loginPost))

	protected := dynamic.Append(app.requireAuth)
	router.Handler(http.MethodPost, "/logout", protected.ThenFunc(app.logout))
	router.Handler(http.MethodGet, "/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/create", protected.ThenFunc(app.snippetCreatePost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)

}
