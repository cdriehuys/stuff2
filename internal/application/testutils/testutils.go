package testutils

import (
	"bytes"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/cdriehuys/stuff2/internal/application"
	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/cdriehuys/stuff2/internal/templating"
	"github.com/cdriehuys/stuff2/translations"
	"github.com/cdriehuys/stuff2/ui"
)

func NewTestApplication(t *testing.T) *application.Application {
	discardLogger := slog.New(slog.DiscardHandler)

	sessionManager := scs.New()
	sessionManager.Cookie.HttpOnly = true

	// Default to using the embedded file system like production.
	templateFS, err := fs.Sub(ui.FS, "templates")
	if err != nil {
		t.Fatalf("failed to load templates from file system: %v", err)
	}

	templates, err := templating.NewTemplateCache(discardLogger, templateFS)
	if err != nil {
		t.Fatalf("failed to construct template cache: %v", err)
	}

	// Load translations
	ut, err := i18n.LoadTranslations(discardLogger, translations.FS)
	if err != nil {
		t.Fatalf("Failed to load translations: %v", err)
	}

	return &application.Application{
		Logger:     discardLogger,
		Session:    sessionManager,
		Templates:  templates,
		Translator: ut,
	}
}

type TestServer struct {
	*httptest.Server
}

func NewTestServer(t *testing.T, h http.Handler) *TestServer {
	ts := httptest.NewServer(h)

	// CSRF protection depends on persisting cookies.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}

	ts.Client().Jar = jar

	// Prevent the test server client from following redirects to allow for testing against the
	// redirect response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &TestServer{ts}
}

type TestResponse struct {
	Status  int
	Headers http.Header
	Cookies []*http.Cookie
	Body    string
}

func MakeTestResponse(t *testing.T, res *http.Response) TestResponse {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return TestResponse{
		Status:  res.StatusCode,
		Headers: res.Header,
		Cookies: res.Cookies(),
		Body:    string(bytes.TrimSpace(body)),
	}
}

func (ts *TestServer) Get(t *testing.T, path string) TestResponse {
	req := ts.makeRequest(t, http.MethodGet, path, nil)

	return ts.doRequest(t, req)
}

func (ts *TestServer) PostForm(t *testing.T, path string, form url.Values) TestResponse {
	req := ts.makeRequest(t, http.MethodPost, path, strings.NewReader(form.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	return ts.doRequest(t, req)
}

func (ts *TestServer) makeRequest(t *testing.T, method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatalf("failed to create %s request to %q: %v", method, path, err)
	}

	return req
}

func (ts *TestServer) doRequest(t *testing.T, req *http.Request) TestResponse {
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("failed to send %s request to %q: %v", req.Method, req.URL.Path, err)
	}

	defer res.Body.Close()

	return MakeTestResponse(t, res)
}
