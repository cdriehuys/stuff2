package application_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application"
	"github.com/cdriehuys/stuff2/internal/application/testutils"
	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/cdriehuys/stuff2/internal/models/mocks"
	ut "github.com/go-playground/universal-translator"
)

func TestApplication_registerGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/register")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
}

type CapturingTemplateEngine[T any] struct {
	RenderError error

	RenderedName    string
	RenderedData    T
	RenderedRawData any
}

func (e *CapturingTemplateEngine[T]) Render(w io.Writer, name string, data any) error {
	e.RenderedName = name
	e.RenderedRawData = data

	if structuredData, ok := data.(T); ok {
		e.RenderedData = structuredData
	}

	// Have to return error before writing, otherwise the response will automatically get a 200
	// response.
	if e.RenderError != nil {
		return e.RenderError
	}

	fmt.Fprintf(w, "%#v", data)

	return nil
}

type WantRedirect struct {
	Status   int
	Location string
}

func TestApplication_registerPost(t *testing.T) {
	defaultEmail := "test@example.com"
	defaultPassword := "tops3cret"

	testCases := []struct {
		name           string
		templates      CapturingTemplateEngine[application.TemplateData]
		users          mocks.UserModel
		email          string
		password       string
		wantStatus     int
		wantRegistered models.NewUser
		wantRedirect   *WantRedirect
	}{
		{
			name:           "successful registration",
			email:          defaultEmail,
			password:       defaultPassword,
			wantRegistered: models.NewUser{Email: defaultEmail, Password: defaultPassword},
			wantRedirect:   &WantRedirect{Status: http.StatusSeeOther, Location: "/register/success"},
		},
		{
			name: "registration server error",
			users: mocks.UserModel{
				RegisterError: errors.New("registration failed"),
			},
			email:          defaultEmail,
			password:       defaultPassword,
			wantRegistered: models.NewUser{Email: defaultEmail, Password: defaultPassword},
			wantStatus:     http.StatusInternalServerError,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)

			app.Templates = &tt.templates
			app.Users = &tt.users

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			form := csrfFormValues(t, app, ts, "/register")
			form.Add("email", tt.email)
			form.Add("password", tt.password)

			res := ts.PostForm(t, "/register", form)

			if tt.wantStatus != 0 && res.Status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}

			if got := tt.users.RegisteredUser.Email; got != tt.wantRegistered.Email {
				t.Errorf("Expected registered user email %q, got %q", tt.wantRegistered.Email, got)
			}

			if got := tt.users.RegisteredUser.Password; got != tt.wantRegistered.Password {
				t.Errorf("Expected registered user password %q, got %q", tt.wantRegistered.Password, got)
			}

			if want := tt.wantRedirect; want != nil {
				if res.Status != want.Status {
					t.Errorf("Expected status %d, got %d", want.Status, res.Status)
				}

				if got := res.Headers.Get("Location"); got != want.Location {
					t.Errorf("Expected redirect location %q, got %q", want.Location, got)
				}
			}
		})
	}
}

type MockValidationError struct {
	field string
	tag   string
	value any
}

func (e MockValidationError) Tag() string {
	return e.tag
}

func (e MockValidationError) ActualTag() string {
	return e.tag
}

func (e MockValidationError) Namespace() string {
	return ""
}

func (e MockValidationError) StructNamespace() string {
	return ""
}

func (e MockValidationError) Field() string {
	return e.field
}

func (e MockValidationError) StructField() string {
	return e.field
}

func (e MockValidationError) Value() any {
	return e.value
}

func (e MockValidationError) Param() string {
	return ""
}

func (e MockValidationError) Kind() reflect.Kind {
	return reflect.String
}

func (e MockValidationError) Type() reflect.Type {
	return nil
}

func (e MockValidationError) Translate(ut.Translator) string {
	return ""
}

func (e MockValidationError) Error() string {
	return e.tag
}

func TestApplication_registerPost_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name              string
		email             string
		password          string
		wantErroredFields []string
	}{
		{
			name:              "fields missing",
			email:             "",
			password:          "",
			wantErroredFields: []string{"email", "password"},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)

			templates := &CapturingTemplateEngine[application.TemplateData]{}
			app.Templates = templates
			app.Users = &mocks.UserModel{}

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			form := csrfFormValues(t, app, ts, "/register")
			form.Add("email", tt.email)
			form.Add("password", tt.password)

			res := ts.PostForm(t, "/register", form)

			if res.Status != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
			}

			formData := templates.RenderedData.Form
			for _, wantErroredField := range tt.wantErroredFields {
				field, exists := formData.Fields[wantErroredField]
				if !exists {
					t.Fatalf("expected field error for %q, but the field is missing from the form", wantErroredField)
				}

				if len(field.Errors.Errors()) == 0 {
					t.Errorf("Expected %q to have errors", wantErroredField)
				}
			}
		})
	}
}

func TestApplication_registerSuccess(t *testing.T) {
	testCases := []struct {
		name       string
		templates  CapturingTemplateEngine[application.TemplateData]
		wantStatus int
	}{
		{
			name:       "successful render",
			wantStatus: http.StatusOK,
		},
		{
			name: "render error",
			templates: CapturingTemplateEngine[application.TemplateData]{
				RenderError: errors.New("rendering failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)
			app.Templates = &tt.templates

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			res := ts.Get(t, "/register/success")

			if res.Status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}
		})
	}
}
