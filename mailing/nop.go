package mailing

import "context"

// nop implements Sender but doesn't interact with any email service provider.
type nop struct{}

// Send doesn't send an email.
func (n *nop) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	return nil
}

// NewNopEmailSender returns a no-op Sender. It never interacts with an email service provider.
// This Sender implementation is useful when a service doesn't want to enable sending emails.
func NewNopEmailSender() Sender {
	return &nop{}
}
