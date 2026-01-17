package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
)

type Emailer interface {
	Send(ctx context.Context, to string, from string, subject string, body string) error
}

type EmailTemplateData struct {
	VerificationLink string
}

type EmailVerifier struct {
	logger *slog.Logger

	emailer   Emailer
	templates TemplateEngine

	baseDomain *url.URL
	sender     string
}

func NewEmailVerifier(logger *slog.Logger, emailer Emailer, templates TemplateEngine, baseDomain *url.URL, sender string) *EmailVerifier {
	return &EmailVerifier{
		logger:     logger,
		emailer:    emailer,
		templates:  templates,
		baseDomain: baseDomain,
		sender:     sender,
	}
}

func (v *EmailVerifier) DuplicateRegistration(ctx context.Context, email string) error {
	body, err := v.render("duplicate-email.txt", EmailTemplateData{})
	if err != nil {
		return fmt.Errorf("rendering duplicate email template: %v", err)
	}

	return v.emailer.Send(ctx, email, v.sender, "Duplicate Registration", body)
}

func (v *EmailVerifier) NewEmail(ctx context.Context, email string, token string) error {
	verificationLink := v.baseDomain.JoinPath("verify-email", token).String()
	data := EmailTemplateData{VerificationLink: verificationLink}

	body, err := v.render("new-registration.txt", data)
	if err != nil {
		return fmt.Errorf("rendering new registration email template: %v", err)
	}

	return v.emailer.Send(ctx, email, v.sender, "Verify Your Email", body)
}

func (v *EmailVerifier) render(subject string, data EmailTemplateData) (string, error) {
	var output strings.Builder
	if err := v.templates.Render(&output, subject, data); err != nil {
		return "", fmt.Errorf("rendering email template %q: %v", subject, err)
	}

	return output.String(), nil
}
