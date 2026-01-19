package application

import (
	"net/http"

	"github.com/justinas/alice"
)

func (a *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	// Middleware applied to dynamic requests, ie requests that depend on the user who sent them.
	dynamic := alice.New(a.preventCSRF)

	mux.Handle("GET /{$}", dynamic.ThenFunc(a.homeGet))
	mux.Handle("GET /login", dynamic.ThenFunc(a.loginGet))
	mux.Handle("POST /login", dynamic.ThenFunc(a.loginPost))
	mux.Handle("GET /register", dynamic.ThenFunc(a.registerGet))
	mux.Handle("POST /register", dynamic.ThenFunc(a.registerPost))
	mux.Handle("GET /register/success", dynamic.ThenFunc(a.registerSuccess))
	mux.Handle("GET /verify-email/{token}", dynamic.ThenFunc(a.verifyEmailGet))
	mux.Handle("POST /verify-email/{token}", dynamic.ThenFunc(a.verifyEmailPost))
	mux.Handle("GET /verify-email-success", dynamic.ThenFunc(a.verifyEmailSuccess))

	// Middleware applied to all requests.
	standard := alice.New(a.RecoverPanic, a.translatorMiddleware)

	return standard.Then(mux)
}
