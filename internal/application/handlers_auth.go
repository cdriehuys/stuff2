package application

import (
	"errors"
	"net/http"

	"github.com/cdriehuys/stuff2/internal/forms"
	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/cdriehuys/stuff2/internal/validation"
)

func (a *Application) loginGet(w http.ResponseWriter, r *http.Request) {
	form := forms.Form{
		Fields: map[string]forms.Field{
			"email":    {Name: "email"},
			"password": {Name: "password"},
		},
	}

	data := a.templateData(r)
	data.Form = form

	a.render(w, r, "login.html", data)
}

func (a *Application) loginPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	_, err := a.Users.Authenticate(r.Context(), email, password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			t := a.translator(r)

			form := forms.Form{
				Errors: []validation.Error{validation.MakeError("credentials", t.T("login.credentials.invalid"))},
				Fields: map[string]forms.Field{
					"email":    {Name: "email", Value: email},
					"password": {Name: "password"},
				},
			}

			data := a.templateData(r)
			data.Form = form

			w.WriteHeader(http.StatusUnauthorized)
			a.render(w, r, "login.html", data)
			return
		}

		a.serverError(w, r, "Failed to authenticate.", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *Application) registerGet(w http.ResponseWriter, r *http.Request) {
	form := forms.Form{
		Fields: map[string]forms.Field{
			"email":    {Name: "email"},
			"password": {Name: "password"},
		},
	}

	data := a.templateData(r)
	data.Form = form

	a.render(w, r, "register.html", data)
}

func (a *Application) registerPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	rawEmail := r.PostFormValue("email")
	rawPassword := r.PostFormValue("password")

	newUser, err := models.MakeNewUser(r.Context(), rawEmail, rawPassword)
	if err != nil {
		email := forms.Field{Name: "email", Value: rawEmail}
		// Don't add the user's plaintext password back to the response
		password := forms.Field{Name: "password"}

		userErrors := models.NewUserErrors{}
		if errors.As(err, &userErrors) {
			email.Errors = userErrors.Email
			password.Errors = userErrors.Password

			form := forms.Form{
				Fields: map[string]forms.Field{
					"email":    email,
					"password": password,
				},
			}

			data := a.templateData(r)
			data.Form = form

			a.render(w, r, "register.html", data)
			return
		}
	}

	if err := a.Users.Register(r.Context(), newUser); err != nil {
		a.serverError(w, r, "Failed to register user.", err)
		return
	}

	http.Redirect(w, r, "/register/success", http.StatusSeeOther)
}

func (a *Application) registerSuccess(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register-success.html", a.templateData(r))
}

func (a *Application) verifyEmailGet(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "verify-email.html", a.templateData(r))
}

func (a *Application) verifyEmailPost(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	if err := a.Users.VerifyEmail(r.Context(), token); err != nil {
		if errors.Is(err, models.ErrInvalidEmailVerificationToken) {
			t := a.translator(r)
			form := forms.Form{
				Errors: []validation.Error{validation.MakeError("invalid", t.T("email.verification.key.invalid"))},
			}

			data := a.templateData(r)
			data.Form = form

			w.WriteHeader(http.StatusBadRequest)
			a.render(w, r, "verify-email.html", data)
			return
		}

		a.serverError(w, r, "Failed to verify email address.", err)
		return
	}

	http.Redirect(w, r, "/verify-email-success", http.StatusSeeOther)
}

func (a *Application) verifyEmailSuccess(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "verify-email-success.html", a.templateData(r))
}
