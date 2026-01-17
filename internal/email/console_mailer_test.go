package email_test

import (
	"strings"
	"testing"

	"github.com/cdriehuys/stuff2/internal/email"
)

func TestConsoleMailer_Send(t *testing.T) {
	testCases := []struct {
		name    string
		to      string
		from    string
		subject string
		body    string
	}{
		{
			name:    "simple message",
			to:      "test@example.com",
			from:    "no-reply@localhost",
			subject: "Test Message",
			body:    "Hello, World!",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var writer strings.Builder
			mailer := email.NewConsoleMailer(&writer)

			if err := mailer.Send(t.Context(), tt.to, tt.from, tt.subject, tt.body); err != nil {
				t.Fatalf("mailer failed: %v", err)
			}

			output := writer.String()

			if !strings.Contains(output, tt.to) {
				t.Errorf("Expected to find %q in output:\n%s", tt.to, output)
			}

			if !strings.Contains(output, tt.from) {
				t.Errorf("Expected to find %q in output:\n%s", tt.from, output)
			}

			if !strings.Contains(output, tt.subject) {
				t.Errorf("Expected to find %q in output:\n%s", tt.subject, output)
			}

			if !strings.Contains(output, tt.body) {
				t.Errorf("Expected to find %q in output:\n%s", tt.body, output)
			}
		})
	}
}
