package application

import (
	"errors"
	"net/http"

	"github.com/cdriehuys/stuff2/internal/forms"
	"github.com/cdriehuys/stuff2/internal/models"
)

func (a *Application) registerGet(w http.ResponseWriter, r *http.Request) {
	form := forms.Form{
		Fields: map[string]forms.Field{
			"email":    {Name: "email"},
			"password": {Name: "password"},
		},
	}

	var data TemplateData = a.templateData(r)
	data.Form = form

	a.render(w, r, "register.html", data)
}

func (a *Application) registerPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	newUser, err := models.MakeNewUser(r.Context(), r.PostFormValue("email"), r.PostFormValue("password"))
	if err != nil {
		email := forms.Field{Name: "email", Value: newUser.Email}
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
