package application

import (
	"fmt"
	"net/http"

	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/justinas/nosurf"
)

func (a *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			panicData := recover()
			if panicData != nil {
				// Guard against issues with a panic in the middle of writing a response which could
				// cause problems if the client tries to reuse the connection for something else.
				w.Header().Set("Connection", "close")

				// Ideally the panic was before attempting to write a response and we can attempt to
				// send a full error page.
				a.serverError(w, r, "Unhandled panic.", fmt.Errorf("%v", panicData))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *Application) preventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})

	return csrfHandler
}

var translatorContextKey = struct{ name string }{name: "translator"}

func (a *Application) translatorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := i18n.NewRequestTranslator(a.Logger, a.Translator, r)

		r = r.WithContext(i18n.AddToContext(r.Context(), t))

		next.ServeHTTP(w, r)
	})
}
