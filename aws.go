package ign

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

// SendEmail using AWS Simple Email Service (SES)
// The following environment variables must be set:
//
// AWS_REGION
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
func SendEmail(sender, recipient, subject, body string) error {

	// The character encoding for the email.
	charSet := "UTF-8"

	sess := session.Must(session.NewSession())

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Attempt to send the email.
	_, err := svc.SendEmail(input)

	// Return error messages if they occur.
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
			return errors.New(code + " " + aerr.Error())
		}
		return errors.New(err.Error())
	}

	return nil
}
