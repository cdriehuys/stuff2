package application_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"testing"

	"github.com/cdriehuys/stuff2/internal/application"
)

const (
	expectedVerificationPathSegment = "verify-email"
)

type mockEmailTemplateEngine struct {
	renderedSubject string
	renderedData    application.EmailTemplateData
	renderedRawData any

	renderError error
}

func (e *mockEmailTemplateEngine) Render(w io.Writer, subject string, data any) error {
	if e.renderError != nil {
		return e.renderError
	}

	e.renderedSubject = subject
	e.renderedRawData = data

	if emailData, ok := data.(application.EmailTemplateData); ok {
		e.renderedData = emailData
	}

	fmt.Fprintf(w, "%#v", data)

	return nil
}

type capturingMailer struct {
	sendTo      string
	sendFrom    string
	sendSubject string
	sendBody    string
	sendError   error
}

func (m *capturingMailer) Send(ctx context.Context, to string, from string, subject string, body string) error {
	m.sendTo = to
	m.sendFrom = from
	m.sendSubject = subject
	m.sendBody = body

	return m.sendError
}

func TestEmailVerifier_DuplicateRegistration(t *testing.T) {
	baseDomain, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("Invalid base domain: %v", err)
	}

	testCases := []struct {
		name             string
		mailer           capturingMailer
		templates        mockEmailTemplateEngine
		sender           string
		email            string
		wantEmailTo      string
		wantEmailFrom    string
		wantEmailSubject string
		wantErr          bool
	}{
		{
			name:             "successful send",
			sender:           "admin@localhost",
			email:            "new-user@example.com",
			wantEmailTo:      "new-user@example.com",
			wantEmailFrom:    "admin@localhost",
			wantEmailSubject: "Duplicate Registration",
		},
		{
			name: "rendering error",
			templates: mockEmailTemplateEngine{
				renderError: errors.New("rendering failed"),
			},
			sender:  "admin@localhost",
			email:   "new-user@example.com",
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			verifier := application.NewEmailVerifier(slog.New(slog.DiscardHandler), &tt.mailer, &tt.templates, baseDomain, tt.sender)

			err = verifier.DuplicateRegistration(t.Context(), tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error presence %v, got error %#v", tt.wantErr, err)
			}

			if got := tt.mailer.sendTo; got != tt.wantEmailTo {
				t.Errorf("Expected email to be sent to %q, got %q", tt.wantEmailTo, got)
			}

			if got := tt.mailer.sendFrom; got != tt.wantEmailFrom {
				t.Errorf("Expected email to be sent from %q, got %q", tt.wantEmailFrom, got)
			}

			if got := tt.mailer.sendSubject; got != tt.wantEmailSubject {
				t.Errorf("Expected email subject %q, got %q", tt.wantEmailSubject, got)
			}
		})
	}
}

func TestEmailVerifier_NewEmail(t *testing.T) {
	testCases := []struct {
		name             string
		mailer           capturingMailer
		templates        mockEmailTemplateEngine
		baseDomain       string
		sender           string
		email            string
		token            string
		wantEmailTo      string
		wantEmailFrom    string
		wantEmailSubject string
		wantEmailToken   string
		wantErr          bool
	}{
		{
			name:             "successful send",
			baseDomain:       "http://localhost",
			sender:           "admin@localhost",
			email:            "new-user@example.com",
			token:            "secret-token",
			wantEmailTo:      "new-user@example.com",
			wantEmailFrom:    "admin@localhost",
			wantEmailSubject: "Verify Your Email",
			wantEmailToken:   "secret-token",
		},
		{
			name: "rendering error",
			templates: mockEmailTemplateEngine{
				renderError: errors.New("rendering failed"),
			},
			baseDomain: "http://localhost",
			sender:     "admin@localhost",
			email:      "new-user@example.com",
			token:      "secret-token",
			wantErr:    true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			baseDomain, err := url.Parse(tt.baseDomain)
			if err != nil {
				t.Fatalf("Base domain %q is invalid: %v", tt.baseDomain, err)
			}

			verifier := application.NewEmailVerifier(slog.New(slog.DiscardHandler), &tt.mailer, &tt.templates, baseDomain, tt.sender)

			err = verifier.NewEmail(t.Context(), tt.email, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error presence %v, got error %#v", tt.wantErr, err)
			}

			if got := tt.mailer.sendTo; got != tt.wantEmailTo {
				t.Errorf("Expected email to be sent to %q, got %q", tt.wantEmailTo, got)
			}

			if got := tt.mailer.sendFrom; got != tt.wantEmailFrom {
				t.Errorf("Expected email to be sent from %q, got %q", tt.wantEmailFrom, got)
			}

			if got := tt.mailer.sendSubject; got != tt.wantEmailSubject {
				t.Errorf("Expected email subject %q, got %q", tt.wantEmailSubject, got)
			}

			if tt.wantEmailToken != "" {
				wantLink := baseDomain.JoinPath(expectedVerificationPathSegment, tt.wantEmailToken).String()
				if got := tt.templates.renderedData.VerificationLink; got != wantLink {
					t.Errorf("Expected verification link %q, got %q", wantLink, got)
				}
			}
		})
	}
}
