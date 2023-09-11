package mailing

import (
	"context"
	"fmt"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
)

// sendgridEmailService implements the Sender interface using Sendgrid.
//
//	Reference: https://github.com/sendgrid/sendgrid-go
type sendgridEmailService struct {
	client sendgridSender
	contentInjector
}

func (s *sendgridEmailService) emailBuilder() sendgridEmailBuilder {
	return newSendgridEmailBuilder()
}

// Send sends an email from sender to the given recipients. The email body is composed by an HTML template
// that is filled in with values provided in data.
func (s *sendgridEmailService) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	err := validateEmail(sender, recipients, cc, bcc, data)
	if err != nil {
		return err
	}

	builder := s.emailBuilder().
		Sender(sender).
		Recipients(recipients).
		CC(cc).
		BCC(bcc).
		Subject(subject)

	builder, err = s.contentInjector(builder, template, data)
	if err != nil {
		return err
	}
	m := builder.Build()

	res, err := s.client.SendWithContext(ctx, m)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("failed to send email (%d) %s: %s", res.StatusCode, http.StatusText(res.StatusCode), res.Body)
	}
	return nil
}

// NewSendgridEmailSender initializes a new Sender with a sendgrid client. It will send emails using Go templates.
// See NewTemplateSender for conveniently handling Go templates in your project.
func NewSendgridEmailSender(client sendgridSender) Sender {
	return newSendgridEmailSender(client, injectHTMLContent)
}

// NewSendgridDynamicTemplatesEmailSender initializes a new Sender with a sendgrid client. It will send emails through
// sendgrid using dynamic templates defined in the Sendgrid dashboard.
func NewSendgridDynamicTemplatesEmailSender(client sendgridSender) Sender {
	return newSendgridEmailSender(client, injectTemplateContent)
}

func newSendgridEmailSender(client sendgridSender, injector contentInjector) Sender {
	return &sendgridEmailService{
		client:          client,
		contentInjector: injector,
	}
}

// parseSendgridEmails converts the given slice of emails to sendgrid emails.
func parseSendgridEmails(emails []string) []*mail.Email {
	out := make([]*mail.Email, len(emails))
	for i := 0; i < len(emails); i++ {
		out[i] = mail.NewEmail("", emails[i])
	}
	return out
}
