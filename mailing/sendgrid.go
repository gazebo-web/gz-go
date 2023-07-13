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

// Send sends an email from sender to the given recipients. The email body is composed by an HTML template
// that is filled in with values provided in data.
func (s *sendgridEmailService) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	err := validateEmail(recipients, sender, data)
	if err != nil {
		return err
	}

	htmlContent, err := gz.ParseHTMLTemplate(template, data)
	if err != nil {
		return err
	}

	m := prepareSendgridMailV3(sender, recipients, subject, htmlContent)

	res, err := s.client.SendWithContext(ctx, m)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email: (%d) %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return nil
}

// prepareSendgridMailV3 prepares the input values used for sending an email.
func prepareSendgridMailV3(sender string, recipients []string, subject string, htmlContent string) *mail.SGMailV3 {
	p := mail.NewPersonalization()
	p.AddFrom(mail.NewEmail("", sender))
	p.AddTos(parseSendgridEmails(recipients)...)
	p.Subject = subject

	m := mail.NewV3Mail()
	m.AddPersonalizations(p)
	m.AddContent(mail.NewContent("text/html", htmlContent))
	return m
}

// NewSendgridEmailSender initializes a new Sender with a sendgrid client.
func NewSendgridEmailSender(client sendgridSender) Sender {
	return &sendgridEmailService{
		client: client,
	}
}

func parseSendgridEmails(emails []string) []*mail.Email {
	out := make([]*mail.Email, len(emails))
	for i := 0; i < len(emails); i++ {
		out[i] = mail.NewEmail("", emails[i])
	}
	return out
}
