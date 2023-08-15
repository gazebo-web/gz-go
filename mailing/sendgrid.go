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
	contentInjector
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
	b.personalization.DynamicTemplateData, err = structs.ToMap(data)
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

// parseSendgridEmails converts the given slice of emails to sendgrid emails.
func parseSendgridEmails(emails []string) []*mail.Email {
	out := make([]*mail.Email, len(emails))
	for i := 0; i < len(emails); i++ {
		out[i] = mail.NewEmail("", emails[i])
	}
	return out
}

// contentInjector defines the function signature that injects content injection into sendgridEmailBuilder
type contentInjector func(b sendgridEmailBuilder, key string, data any) (sendgridEmailBuilder, error)

// injectTemplateContent injects the information needed to send a Sendgrid email using dynamic templates.
func injectTemplateContent(b sendgridEmailBuilder, id string, data any) (sendgridEmailBuilder, error) {
	return b.Template(id, data), nil
}

// injectHTMLContent injects the actual content after parsing an HTML Go template.
func injectHTMLContent(b sendgridEmailBuilder, path string, data any) (sendgridEmailBuilder, error) {
	content, err := gz.ParseHTMLTemplate(path, data)
	if err != nil {
		return b, err
	}
	return b.Content("text/html", content), nil
}

// NewSendgridEmailSender initializes a new Sender with a sendgrid client. It will send emails using Go templates.
// See NewTemplateSender for conveniently handling Go templates in your project.
func NewSendgridEmailSender(client sendgridSender) Sender {
	return &sendgridEmailService{
		client:          client,
		contentInjector: injectHTMLContent,
	}
}

// NewSendgridDynamicTemplatesEmailSender initializes a new Sender with a sendgrid client. It will send emails through
// sendgrid using dynamic templates defined in the Sendgrid dashboard.
func NewSendgridDynamicTemplatesEmailSender(client sendgridSender) Sender {
	return &sendgridEmailService{
		client:          client,
		contentInjector: injectTemplateContent,
	}
}
