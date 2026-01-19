package application_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application/testutils"
	"github.com/google/uuid"
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

type mockSessionManager struct {
	data map[string]any
}

func (m *mockSessionManager) Get(_ context.Context, key string) any {
	return m.data[key]
}

func (m *mockSessionManager) LoadAndSave(next http.Handler) http.Handler {
	if m.data == nil {
		m.data = make(map[string]any)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (m *mockSessionManager) Put(_ context.Context, key string, value any) {
	m.data[key] = value
}

func TestApplication_RequireAuthenticated(t *testing.T) {
	userID := uuid.New()

	testCases := []struct {
		name           string
		sessionManager mockSessionManager
		wantStatus     int
		wantLocation   string
	}{
		{
			name:         "not authenticated",
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/login",
		},
		{
			name: "malformed auth data",
			sessionManager: mockSessionManager{
				data: map[string]any{
					"user_id": "lizard",
				},
			},
			wantStatus:   http.StatusSeeOther,
			wantLocation: "/login",
		},
		{
			name: "authenticated",
			sessionManager: mockSessionManager{
				data: map[string]any{
					"user_id": userID.String(),
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			app := testutils.NewTestApplication(t)
			app.Session = &tt.sessionManager

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := tt.sessionManager.LoadAndSave(app.RequireAuthenticated(handler))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/some/protected/route", nil)

			wrapped.ServeHTTP(w, r)

			res := w.Result()
			if res.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, res.StatusCode)
			}

			if got := res.Header.Get("Location"); tt.wantLocation != got {
				t.Errorf("Expected Location header %q, got %q", tt.wantLocation, got)
			}
		})
	}
}
