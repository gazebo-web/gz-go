package mailing

import "github.com/gazebo-web/gz-go/v8"

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
