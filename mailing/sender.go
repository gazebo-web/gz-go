package mailing

import (
	"context"
	"errors"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// charset is the default email encoding.
const charset = "UTF-8"

var (
	// ErrEmptyRecipientList is returned when an empty recipient list is passed to the Sender.Send method.
	ErrEmptyRecipientList = errors.New("empty recipient list")
	// ErrEmptySender is returned when an empty sender email address is passed to the Sender.Send method.
	ErrEmptySender = errors.New("empty sender")
	// ErrInvalidSender is returned when an invalid email address is passed to the Sender.Send method.
	ErrInvalidSender = errors.New("invalid sender")
	// ErrInvalidRecipient is returned when an invalid email is passed in the list of recipients to the Sender.Send method.
	ErrInvalidRecipient = errors.New("invalid recipient")
	// ErrInvalidData is returned when an invalid data is passed to the Sender.Send method.
	ErrInvalidData = errors.New("invalid data")
)

// Sender allows sending emails through an email service.
type Sender interface {
	// Send sends an email from sender to the given recipients. The email body is described by a template
	// that contains the information passed through data.
	Send(ctx context.Context, recipients []string, sender, subject, template string, data any) error
}

type sendgridSender interface {
	SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error)
}
