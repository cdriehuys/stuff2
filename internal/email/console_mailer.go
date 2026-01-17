package email

import (
	"context"
	"fmt"
	"io"
	"strings"
)

var (
	messageSeparator = strings.Repeat("*", 80)
	bodySeparator    = strings.Repeat("-", 80)
)

// ConsoleMailer writes emails to the given output stream for use in development.
type ConsoleMailer struct {
	w io.Writer
}

// NewConsoleMailer creates a console mailer that writes to the given writer.
func NewConsoleMailer(w io.Writer) *ConsoleMailer {
	return &ConsoleMailer{w}
}

// Send writes the contents of the email to the mailer's output.
func (m *ConsoleMailer) Send(ctx context.Context, to string, from string, subject string, body string) error {
	fmt.Fprintf(m.w, "\n\n%s\n", messageSeparator)
	fmt.Fprintf(m.w, "To: %s\n", to)
	fmt.Fprintf(m.w, "From: %s\n", from)
	fmt.Fprintf(m.w, "Subject: %s\n", subject)
	fmt.Fprintln(m.w, bodySeparator)
	fmt.Fprintln(m.w, body)
	fmt.Fprintln(m.w, messageSeparator)

	return nil
}
