package mailing

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/gazebo-web/gz-go/v8"
)

// awsSimpleEmailService implements the Sender interface using AWS Simple Email Service API.
// This implementation uses the AWS SDK v1.
//
//	References:
//	- AWS SDK V1: https://github.com/aws/aws-sdk-go
//	- AWS SES SDK V1: https://pkg.go.dev/github.com/aws/aws-sdk-go/service/ses
type awsSimpleEmailService struct {
	API sesiface.SESAPI
}

// Send sends an email from sender to the given recipients. The email body is composed by an HTML template
// that is filled in with values provided in data.
func (e *awsSimpleEmailService) Send(ctx context.Context, sender string, recipients, cc, bcc []string, subject, template string, data any) error {
	err := validEmail(recipients, sender, data)
	if err != nil {
		return err
	}

	content, err := gz.ParseHTMLTemplate(template, data)
	if err != nil {
		return err
	}

	err = e.send(ctx, sender, recipients, subject, content)
	if err != nil {
		return err
	}

	return nil
}

// send attempts to send an email using the AWS SES service.
func (e *awsSimpleEmailService) send(_ context.Context, sender string, recipients []string, subject string, content string) error {
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

	// Attempt to send the awsSimpleEmailService.
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
	return &awsSimpleEmailService{
		API: api,
	}
}
