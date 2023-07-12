package mailing

import (
	"context"
	"fmt"
	"github.com/gazebo-web/gz-go/v8"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
)

// sendgridEmailService implements the Sender interface using Sendgrid.
//
//	Reference: https://github.com/sendgrid/sendgrid-go
type sendgridEmailService struct {
	client sendgridSender
}

// Send sends an email from sender to the given recipients. The email body is described by a template
// that contains the information passed through data.
func (s *sendgridEmailService) Send(ctx context.Context, recipients []string, sender, subject, template string, data any) error {
	from := mail.NewEmail("", sender)
	htmlContent, err := gz.ParseHTMLTemplate(template, data)
	if err != nil {
		return err
	}
	for _, recipient := range recipients {
		to := mail.NewEmail("", recipient)
		m := mail.NewSingleEmail(from, subject, to, "", htmlContent)
		res, err := s.client.SendWithContext(ctx, m)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to send email: (%d) %s", res.StatusCode, http.StatusText(res.StatusCode))
		}
	}
	return nil
}

// NewSendgridEmailSender initializes a new Sender with a sendgrid client.
func NewSendgridEmailSender(client sendgridSender) Sender {
	return &sendgridEmailService{
		client: client,
	}
}
