package application_test

import (
	"net/http"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application/testutils"
)

func TestApplication_homeGet(t *testing.T) {
	app := testutils.NewTestApplication(t)
	ts := testutils.NewTestServer(t, app.Routes())
	defer ts.Close()

	res := ts.Get(t, "")
	if res.Status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.Status)
	}

	if res.Body == "" {
		t.Error("Expected non-empty body.")
	}
}
