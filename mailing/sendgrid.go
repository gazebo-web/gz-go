package mailing

import (
	"context"
	"fmt"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
	"strings"
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
	if res.StatusCode != http.StatusOK {
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

// parseSendgridEmails converts the given slice of recipients to sendgrid emails.
// Each recipient must be in the following format: "FirstName LastName:email@example.org".
func parseSendgridEmails(recipients []string) []*mail.Email {
	out := make([]*mail.Email, len(recipients))
	for i := 0; i < len(recipients); i++ {
		name, addr := parseRecipient(recipients[i])
		if len(name) == 0 || len(addr) == 0 {
			continue
		}
		out[i] = mail.NewEmail(name, addr)
	}
	return out
}

func parseRecipient(recipient string) (name string, address string) {
	params := strings.Split(recipient, ":")
	if len(params) != 2 {
		return "", ""
	}
	name = params[0]
	address = params[1]
	return
}
