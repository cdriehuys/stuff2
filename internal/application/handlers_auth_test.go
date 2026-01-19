package application_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application"
	"github.com/cdriehuys/stuff2/internal/application/testutils"
	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/cdriehuys/stuff2/internal/models/mocks"
	"github.com/cdriehuys/stuff2/internal/validation"
	"github.com/google/uuid"
)

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

func TestApplication_loginGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/login")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
}

func TestApplication_loginPost(t *testing.T) {
	defaultUserID := uuid.New()

	testCases := []struct {
		name         string
		users        mocks.UserModel
		templates    CapturingTemplateEngine[application.TemplateData]
		email        string
		password     string
		wantEmail    string
		wantPassword string
		wantStatus   int
		wantRedirect *WantRedirect
	}{
		{
			name: "invalid credentials",
			users: mocks.UserModel{
				AuthenticateError: models.ErrInvalidCredentials,
			},
			email:        "test@example.com",
			password:     "bad-password",
			wantEmail:    "test@example.com",
			wantPassword: "bad-password",
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name: "auth error",
			users: mocks.UserModel{
				AuthenticateError: errors.New("broken"),
			},
			email:        "test@example.com",
			password:     "bad-password",
			wantEmail:    "test@example.com",
			wantPassword: "bad-password",
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name: "valid credentials",
			users: mocks.UserModel{
				AuthenticateUser: models.User{ID: defaultUserID},
			},
			wantRedirect: &WantRedirect{
				Status:   http.StatusSeeOther,
				Location: "/",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)
			app.Templates = &tt.templates
			app.Users = &tt.users

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			form := csrfFormValues(t, app, ts, "/login")
			form.Add("email", tt.email)
			form.Add("password", tt.password)

			res := ts.PostForm(t, "/login", form)

			if tt.wantStatus != 0 && tt.wantStatus != res.Status {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}

			if got := tt.users.AuthenticatedEmail; tt.wantEmail != got {
				t.Errorf("Expected authenticated email %q, got %q", tt.email, got)
			}

			if got := tt.users.AuthenticatedPassword; tt.wantPassword != got {
				t.Errorf("Expected authenticated password %q, got %q", tt.password, got)
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

func TestApplication_registerGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/register")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
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

				if len(field.Errors) == 0 {
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

func TestApplication_verifyEmailGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/verify-email/some-token")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
}

func TestApplication_verifyEmailPost(t *testing.T) {
	testCases := []struct {
		name          string
		templates     CapturingTemplateEngine[application.TemplateData]
		users         mocks.UserModel
		token         string
		wantErrorCode string
		wantStatus    int
		wantRedirect  *WantRedirect
	}{
		{
			name: "verification invalid",
			users: mocks.UserModel{
				VerifyEmailError: models.ErrInvalidEmailVerificationToken,
			},
			token:         "invalid",
			wantErrorCode: "invalid",
			wantStatus:    http.StatusBadRequest,
		},
		{
			name: "verification error",
			users: mocks.UserModel{
				VerifyEmailError: errors.New("everything broke"),
			},
			token:      "valid",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "success",
			token: "valid",
			wantRedirect: &WantRedirect{
				Status:   http.StatusSeeOther,
				Location: "/verify-email-success",
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)

			app.Templates = &tt.templates
			app.Users = &tt.users

			ts := testutils.NewTestServer(t, app.Routes())
			defer ts.Close()

			form := csrfFormValues(t, app, ts, "/verify-email/"+tt.token)

			res := ts.PostForm(t, "/verify-email/"+tt.token, form)

			if tt.wantStatus != 0 && res.Status != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.Status)
			}

			if tt.wantErrorCode != "" {
				if !slices.ContainsFunc(tt.templates.RenderedData.Form.Errors, func(e validation.Error) bool { return e.Code() == tt.wantErrorCode }) {
					t.Errorf("Expected error with code %q, got errors %v", tt.wantErrorCode, tt.templates.RenderedData.Form.Errors)
				}
			}

			if got := tt.users.VerifyEmailToken; got != tt.token {
				t.Errorf("Expected verified token %q, got %q", tt.token, got)
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

func TestApplication_verifyEmailSuccess(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "/verify-email-success")

	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}
}
