package ports

import "context"

// Mailer sends transactional email (magic-link login).
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}
