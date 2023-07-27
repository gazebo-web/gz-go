package mailing

import (
	"context"
	"errors"
)

// ErrTemplateNotFound is returned when a given template is not found inside the list of available templates.
var ErrTemplateNotFound = errors.New("template not found")

// templateWrapper implements Sender allowing the template management to be handled by an inner map between
// a key and a template location.
type templateWrapper struct {
	templates map[string]string
	sender    Sender
}

// Send validates that the given template is contained inside the list of available templates and sends the email.
func (t *templateWrapper) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	value, ok := t.templates[template]
	if !ok {
		return ErrTemplateNotFound
	}
	return t.sender.Send(ctx, sender, recipients, cc, bcc, subject, value, data)
}

// NewTemplateWrapper initializes a new Sender implementation that conveniently handles templates by configuring
// the different available email templates in the given templates argument. It relies on the given sender to actually
// send the email. This wrapper can be used alongside the AWS SES and Sendgrid implementations.
//
//	templates: map[string]string{
//		"event.users.signup": "./templates/users/signup.gohtml",
//	}
func NewTemplateWrapper(sender Sender, templates map[string]string) Sender {
	return &templateWrapper{
		sender:    sender,
		templates: templates,
	}
}
