package mailing

import "github.com/gazebo-web/gz-go/v9"

// contentInjector is a function that prepares and configures email content to be sent by sendgridEmailService.
type contentInjector func(b sendgridEmailBuilder, key string, data any) (sendgridEmailBuilder, error)

// injectTemplateContent configures a SendGrid email to send a dynamic template with the provided data.
// key is the unique SendGrid template identifier.
// data is a map[string]any containing data for template fields.
func injectTemplateContent(b sendgridEmailBuilder, id string, data any) (sendgridEmailBuilder, error) {
	return b.Template(id, data), nil
}

// injectHTMLContent configures a SendGrid email to send a rendered HTML Go template.
// path contains the path to the Go template to render.
// data contains values for template fields.
func injectHTMLContent(b sendgridEmailBuilder, path string, data any) (sendgridEmailBuilder, error) {
	content, err := gz.ParseHTMLTemplate(path, data)
	if err != nil {
		return b, err
	}
	return b.Content("text/html", content), nil
}
