package mailing

import (
	"github.com/gazebo-web/gz-go/v8/structs"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sendgridEmailBuilder allows building mail.SGMailV3 emails.
type sendgridEmailBuilder struct {
	personalization *mail.Personalization
	mail            *mail.SGMailV3
}

// Sender sets the email address where the resulting email comes from.
func (b sendgridEmailBuilder) Sender(from string) sendgridEmailBuilder {
	e := mail.NewEmail("", from)
	b.mail.SetFrom(e)
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

func newSendgridEmailBuilder() sendgridEmailBuilder {
	return sendgridEmailBuilder{
		personalization: mail.NewPersonalization(),
		mail:            mail.NewV3Mail(),
	}
}
