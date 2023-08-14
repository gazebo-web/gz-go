package mailing

import (
	"context"
	"fmt"
	"github.com/gazebo-web/gz-go/v8"
	"github.com/gazebo-web/gz-go/v8/structs"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/http"
)

// sendgridEmailService implements the Sender interface using Sendgrid.
//
//	Reference: https://github.com/sendgrid/sendgrid-go
type sendgridEmailService struct {
	client sendgridSender
}

func (s *sendgridEmailService) emailBuilder() sendgridEmailBuilder {
	return sendgridEmailBuilder{
		personalization: mail.NewPersonalization(),
		mail:            mail.NewV3Mail(),
	}
}

// Send sends an email from sender to the given recipients. The email body is composed by an HTML template
// that is filled in with values provided in data.
func (s *sendgridEmailService) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	err := validateEmail(sender, recipients, cc, bcc, data)
	if err != nil {
		return err
	}

	htmlContent, err := gz.ParseHTMLTemplate(template, data)
	if err != nil {
		return err
	}

	m := s.emailBuilder().
		Sender(sender).
		Recipients(recipients).
		CC(cc).
		BCC(bcc).
		Subject(subject).
		Content("text/html", htmlContent).
		Build()

	res, err := s.client.SendWithContext(ctx, m)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email: (%d) %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	return nil
}

// sendgridEmailBuilder allows building mail.SGMailV3 emails.
type sendgridEmailBuilder struct {
	personalization *mail.Personalization
	mail            *mail.SGMailV3
}

// Sender sets the email address where the resulting email comes from.
func (b sendgridEmailBuilder) Sender(from string) sendgridEmailBuilder {
	b.personalization.AddFrom(mail.NewEmail("", from))
	return b
}

// Recipients sets the email addresses where the resulting email will be sent to.
func (b sendgridEmailBuilder) Recipients(to []string) sendgridEmailBuilder {
	b.personalization.AddTos(parseSendgridEmails(to)...)
	return b
}

// CC sets the email addresses where the resulting email will be carbon-copied to.
func (b sendgridEmailBuilder) CC(ccs []string) sendgridEmailBuilder {
	b.personalization.AddCCs(parseSendgridEmails(ccs)...)
	return b
}

// BCC sets the email addresses where the resulting email will blind carbon-copied to.
func (b sendgridEmailBuilder) BCC(bccs []string) sendgridEmailBuilder {
	b.personalization.AddBCCs(parseSendgridEmails(bccs)...)
	return b
}

// Subject sets the resulting email's subject.
func (b sendgridEmailBuilder) Subject(subject string) sendgridEmailBuilder {
	b.personalization.Subject = subject
	return b
}

// Content sets the resulting email's body with the respective content type.
// It's mutually exclusive with Template.
func (b sendgridEmailBuilder) Content(contentType string, content string) sendgridEmailBuilder {
	b.mail.AddContent(mail.NewContent(contentType, content))
	return b
}

// Template defines the dynamic template identified by ID, and passes the given data as the template
// parameters.
func (b sendgridEmailBuilder) Template(id string, data any) sendgridEmailBuilder {
	var err error
	b.personalization.DynamicTemplateData, err = structs.Map(data, "sendgrid")
	if err != nil {
		return b
	}
	b.mail.SetTemplateID(id)
	return b
}

// Build creates the email from all the parameters previously used.
func (b sendgridEmailBuilder) Build() *mail.SGMailV3 {
	b.mail.AddPersonalizations(b.personalization)
	return b.mail
}

// prepareSendgridMailV3 prepares the input values used for sending an email.
func prepareSendgridMailV3(sender string, recipients []string, cc []string, bcc []string, subject string, htmlContent string) *mail.SGMailV3 {
	p := mail.NewPersonalization()
	p.AddFrom(mail.NewEmail("", sender))
	p.AddTos(parseSendgridEmails(recipients)...)
	p.AddCCs(parseSendgridEmails(cc)...)
	p.AddBCCs(parseSendgridEmails(bcc)...)
	p.Subject = subject

	m := mail.NewV3Mail()
	m.AddPersonalizations(p)
	m.AddContent(mail.NewContent("text/html", htmlContent))
	return m
}

func parseSendgridEmails(emails []string) []*mail.Email {
	out := make([]*mail.Email, len(emails))
	for i := 0; i < len(emails); i++ {
		out[i] = mail.NewEmail("", emails[i])
	}
	return out
}

// NewSendgridEmailSender initializes a new Sender with a sendgrid client.
func NewSendgridEmailSender(client sendgridSender) Sender {
	return &sendgridEmailService{
		client: client,
	}
}
