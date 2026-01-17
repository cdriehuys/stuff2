package application_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application/testutils"
)

func TestApplication_RecoverPanic(t *testing.T) {
	app := testutils.NewTestApplication(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("AHHHHHHHHH!")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/some/endpoint", nil)

	app.RecoverPanic(handler).ServeHTTP(w, r)

	res := w.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}
}
