package mailing

import (
	"context"
	"errors"
	"github.com/gazebo-web/gz-go/v8"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

func TestSendgrid(t *testing.T) {
	suite.Run(t, new(SendgridTestSuite))
}

type SendgridTestSuite struct {
	suite.Suite
	client      sendgridMock
	emailSender Sender
}

func (suite *SendgridTestSuite) SetupTest() {
	suite.client = sendgridMock{}
	suite.emailSender = NewSendgridEmailSender(&suite.client)
	suite.Require().NotNil(suite.emailSender)
}

func (suite *SendgridTestSuite) TestSendEmail_NoRecipients() {
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, []string{}, nil, nil, subject, templatePath, data))

	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_InvalidSenderEmail() {
	const sender = "test.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, []string{}, nil, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_RecipientEmail() {
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	recipients := []string{"test.org"}
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_Error() {
	recipients := []string{"test2@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, recipients, subject, htmlContent)

	expectedError := errors.New("sendgrid failure")
	suite.client.On("SendWithContext", ctx, m).Return((*rest.Response)(nil), expectedError)

	err = suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data)
	suite.Require().Error(err)
	suite.Assert().ErrorIs(err, expectedError)

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_StatusCode() {
	recipients := []string{"test2@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, recipients, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusServiceUnavailable}, error(nil))
	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_Success() {
	recipients := []string{"test2@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, recipients, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_MultipleEmails_Success() {
	recipients := []string{"test2@gazebosim.org", "test3@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, recipients, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestParseEmail() {
	recipients := []string{"test2@gazebosim.org", "test3@gazebosim.org"}

	expected := []*mail.Email{
		mail.NewEmail("", recipients[0]),
		mail.NewEmail("", recipients[1]),
	}

	result := parseSendgridEmails(recipients)

	suite.Assert().Equal(expected, result)
}

type sendgridMock struct {
	mock.Mock
}

func (s *sendgridMock) SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error) {
	args := s.Called(ctx, email)
	return args.Get(0).(*rest.Response), args.Error(1)
}
