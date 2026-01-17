package application_test

import (
	"net/url"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application"
	"github.com/cdriehuys/stuff2/internal/application/testutils"
)

func csrfFormValues(t *testing.T, app *application.Application, ts *testutils.TestServer, formURL string) url.Values {
	templateEngine := app.Templates
	defer func() {
		app.Templates = templateEngine
	}()

	capturer := CapturingTemplateEngine[application.TemplateData]{}
	app.Templates = &capturer

	ts.Get(t, formURL)
	csrfToken := capturer.RenderedData.CSRFToken

	form := url.Values{}
	form.Add("csrf_token", csrfToken)

	return form
}
