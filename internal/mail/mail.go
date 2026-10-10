// Package mail sends the transactional email of the optional login (magic links).
package mail

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/gayratjon-02/Pomosky/internal/ports"
)

// LogMailer writes the email to the log instead of sending it. It is for local
// development only: the log then contains the sign-in link.
type LogMailer struct{ Log *slog.Logger }

var _ ports.Mailer = LogMailer{}

// Send implements ports.Mailer.
func (m LogMailer) Send(_ context.Context, to, subject, body string) error {
	m.Log.Info("email (development, not sent)", "to", to, "subject", subject, "body", body)
	return nil
}

// SMTPMailer sends email through an SMTP server with STARTTLS (port 587).
type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

var _ ports.Mailer = (*SMTPMailer)(nil)

// Send implements ports.Mailer.
func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	msg, err := buildMessage(m.From, to, subject, body, time.Now())
	if err != nil {
		return err
	}
	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}
	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(fmt.Sprintf("%s:%d", m.Host, m.Port), auth, m.From, []string{to}, msg)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func buildMessage(from, to, subject, body string, now time.Time) ([]byte, error) {
	for _, h := range []string{from, to, subject} {
		if strings.ContainsAny(h, "\r\n") || h == "" {
			return nil, errors.New("mail: invalid header value")
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "Date: %s\r\n", now.Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))
	return []byte(b.String()), nil
}
