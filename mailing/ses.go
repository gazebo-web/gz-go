package mailing

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/gazebo-web/gz-go/v8"
)

// simpleEmailService implements the Sender interface using AWS Simple Email Service API.
type simpleEmailService struct {
	API sesiface.SESAPI
}

// Send sends an email to the given recipients from the given sender.
// A template will be parsed with the given data in order to fill the email's body.
// It returns an error when validation fails or sending an email fails.
func (e *simpleEmailService) Send(recipients []string, sender, subject, template string, data interface{}) error {
	err := validEmail(recipients, sender, data)
	if err != nil {
		return err
	}

	content, err := gz.ParseHTMLTemplate(template, data)
	if err != nil {
		return err
	}

	err = e.send(sender, recipients, subject, content)
	if err != nil {
		return err
	}

	return nil
}

// send attempts to send an email using the AWS SES service.
func (e *simpleEmailService) send(sender string, recipients []string, subject string, content string) error {
	input := ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: aws.StringSlice(recipients),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charset),
					Data:    aws.String(content),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charset),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Attempt to send the simpleEmailService.
	_, err := e.API.SendEmail(&input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			var code string
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				code = ses.ErrCodeMessageRejected
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				code = ses.ErrCodeMailFromDomainNotVerifiedException
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				code = ses.ErrCodeConfigurationSetDoesNotExistException
			default:
				code = "Unknown AWS SES error"
			}
			return fmt.Errorf("%s %s", code, aerr.Error())
		}
		return errors.New(err.Error())
	}

	return nil
}

// NewSimpleEmailServiceSender returns a Sender implementation using AWS Simple Email Service.
func NewSimpleEmailServiceSender(api sesiface.SESAPI) Sender {
	return &simpleEmailService{
		API: api,
	}
}
