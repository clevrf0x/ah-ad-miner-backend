package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)
	mux.Use(app.authenticate)

	mux.Get("/status", app.statusHandler)
	// User registration disabled due to security implications
	// mux.Post("/users", app.createUserHandler)
	mux.Post("/authentication-tokens", app.createAuthenticationTokenHandler)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

		mux.Get("/protected", app.protectedTestHandler)

		// Bloodhound data processing
		mux.Get("/results/{id}", app.readResultHandler)
		mux.Post("/results", app.processResultHandler)
	})

	return mux
}
