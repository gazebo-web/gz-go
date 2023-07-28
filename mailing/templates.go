package mailing

import (
	"context"
	"errors"
)

// ErrTemplateNotFound is returned when a given template is not found inside the list of available templates.
var ErrTemplateNotFound = errors.New("template not found")

// templateSender implements Sender allowing the template management to be handled by an inner map between
// a key and a template location.
type templateSender struct {
	templates map[string]string
	sender    Sender
}

// Send validates that the given template is contained inside the list of available templates and sends the email.
func (t *templateSender) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	value, ok := t.templates[template]
	if !ok {
		return ErrTemplateNotFound
	}
	return t.sender.Send(ctx, sender, recipients, cc, bcc, subject, value, data)
}

// NewTemplateSender initializes a new Sender implementation that sends pre-defined templates using another email
// sender. Emails are sent by specifying the template to use and providing data for them.
//
// The sender argument contains a Sender implementation. This is what is actually used to send emails.
//
// The templates argument maps template identifiers to template filepaths. The identifiers are used to select the
// template to send.
//
//	Example:
//	sender := NewSendgridEmailSender(..)
//	sender = NewTemplateSender(sender, map[string]string{ "event.users.signup": "./templates/users/signup.gohtml", })
func NewTemplateSender(sender Sender, templates map[string]string) Sender {
	return &templateSender{
		sender:    sender,
		templates: templates,
	}
}
