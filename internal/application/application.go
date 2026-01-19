package application

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/stuff2/internal/forms"
	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/cdriehuys/stuff2/internal/models"
	ut "github.com/go-playground/universal-translator"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

type SessionManager interface {
	Get(ctx context.Context, key string) any
	LoadAndSave(http.Handler) http.Handler
	Put(ctx context.Context, key string, value any)
}

type TemplateEngine interface {
	Render(io.Writer, string, any) error
}

type UserModel interface {
	Authenticate(ctx context.Context, email string, password string) (models.User, error)
	Register(context.Context, models.NewUser) error
	VerifyEmail(ctx context.Context, token string) error
}

type TemplateData struct {
	IsAuthenticated bool
	CSRFToken       string

	Translator i18n.Translator

	Form forms.Form
}

type Application struct {
	Logger *slog.Logger

	Session    SessionManager
	Templates  TemplateEngine
	Translator *ut.UniversalTranslator

	Users UserModel
}

func (a *Application) translator(r *http.Request) i18n.Translator {
	return i18n.FromContext(r.Context())
}

func (a *Application) templateData(r *http.Request) TemplateData {
	data := TemplateData{
		CSRFToken: nosurf.Token(r),
	}

	if t, ok := r.Context().Value(translatorContextKey).(i18n.Translator); ok {
		data.Translator = t
	}

	return data
}

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, message string, err error, attrs ...any) {
	attrs = append(attrs, "error", err)
	a.Logger.ErrorContext(r.Context(), message, attrs...)

	w.WriteHeader(http.StatusInternalServerError)

	// Calling `a.render` would enter an infinite loop if template rendering is panicking. Attempt
	// the same render process with no error handling instead.
	data := a.templateData(r)
	a.Templates.Render(w, "500.html", data)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, page string, data TemplateData) {
	if err := a.Templates.Render(w, page, data); err != nil {
		a.serverError(w, r, "Failed to render page.", err, "page", page)
	}
}

func (a *Application) homeGet(w http.ResponseWriter, r *http.Request) {
	var data TemplateData = a.templateData(r)
	a.render(w, r, "home.html", data)
}

const sessionKeyUserID = "user_id"

func (a *Application) setAuthenticatedUser(r *http.Request, userID uuid.UUID) {
	a.Session.Put(r.Context(), sessionKeyUserID, userID.String())
}

func (a *Application) getAuthenticatedUserID(r *http.Request) uuid.UUID {
	rawID, ok := a.Session.Get(r.Context(), sessionKeyUserID).(string)
	if !ok {
		return uuid.Nil
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil
	}

	return id
}

func (a *Application) isAuthenticated(r *http.Request) bool {
	return a.getAuthenticatedUserID(r) != uuid.Nil
}
