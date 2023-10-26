package mailing

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gazebo-web/gz-go/v9"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSendgrid(t *testing.T) {
	suite.Run(t, new(SendgridTestSuite))
}

type SendgridTestSuite struct {
	suite.Suite
	client      sendgridMock
	emailSender Sender
	builder     sendgridEmailBuilder
}

func (suite *SendgridTestSuite) SetupTest() {
	suite.client = sendgridMock{}
	suite.emailSender = NewSendgridEmailSender(&suite.client)
	suite.Require().NotNil(suite.emailSender)
	suite.builder = newSendgridEmailBuilder()
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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Content("text/html", htmlContent).
		Build()

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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Content("text/html", htmlContent).
		Build()

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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		Content("text/html", htmlContent).
		Build()

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_CCEmail() {
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	recipients := []string{"test@gazebosim.org"}
	cc := []string{"test.org"}
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		CC(cc).
		Content("text/html", htmlContent).
		Build()

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, cc, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_BCCEmail() {
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	recipients := []string{"test@gazebosim.org"}
	bcc := []string{"test.org"}
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		BCC(bcc).
		Content("text/html", htmlContent).
		Build()

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, bcc, subject, templatePath, data))
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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		Content("text/html", htmlContent).
		Build()

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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		Content("text/html", htmlContent).
		Build()

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

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		Content("text/html", htmlContent).
		Build()

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_MultipleEmails_Success() {
	recipients := []string{"test2@gazebosim.org", "test3@gazebosim.org"}
	cc := []string{"test4@gazebosim.org", "test5@gazebosim.org"}
	bcc := []string{"test6@gazebosim.org", "test7@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		CC(cc).
		BCC(bcc).
		Content("text/html", htmlContent).
		Build()

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, cc, bcc, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridTestSuite) TestSendEmail_WithDynamicTemplates() {
	recipients := []string{"test2@gazebosim.org", "test3@gazebosim.org"}
	cc := []string{"test4@gazebosim.org", "test5@gazebosim.org"}
	bcc := []string{"test6@gazebosim.org", "test7@gazebosim.org"}
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	const templateID = "template-id-123456789"
	ctx := context.Background()

	type emailData struct {
		Test string `structs:"test"`
	}

	data := emailData{Test: "Hello there!"}

	suite.emailSender = newSendgridEmailSender(&suite.client, injectTemplateContent)

	m := suite.builder.
		Sender(sender).
		Subject(subject).
		Recipients(recipients).
		CC(cc).
		BCC(bcc).
		Template(templateID, data).
		Build()

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, cc, bcc, subject, templateID, data))

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
