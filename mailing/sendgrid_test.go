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
	"os"
	"path/filepath"
	"testing"
)

func TestSendgrid(t *testing.T) {
	suite.Run(t, new(SendgridGoTemplatesTestSuite))
}

type SendgridGoTemplatesTestSuite struct {
	suite.Suite
	client      sendgridMock
	emailSender Sender
}

func (suite *SendgridGoTemplatesTestSuite) SetupTest() {
	suite.client = sendgridMock{}
	suite.emailSender = NewSendgridEmailSender(&suite.client)
	suite.Require().NotNil(suite.emailSender)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_NoRecipients() {
	const sender = "test@gazebosim.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, nil, nil, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, []string{}, nil, nil, subject, templatePath, data))

	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_InvalidSenderEmail() {
	const sender = "test.org"
	const subject = "Test email"
	ctx := context.Background()

	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3(sender, nil, nil, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, []string{}, nil, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_RecipientEmail() {
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

	m := prepareSendgridMailV3(sender, nil, nil, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_CCEmail() {
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

	m := prepareSendgridMailV3(sender, recipients, cc, nil, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, cc, nil, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_BCCEmail() {
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

	m := prepareSendgridMailV3(sender, recipients, nil, bcc, subject, htmlContent)

	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, bcc, subject, templatePath, data))
	suite.client.AssertNotCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_Error() {
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

	m := prepareSendgridMailV3(sender, recipients, nil, nil, subject, htmlContent)

	expectedError := errors.New("sendgrid failure")
	suite.client.On("SendWithContext", ctx, m).Return((*rest.Response)(nil), expectedError)

	err = suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data)
	suite.Require().Error(err)
	suite.Assert().ErrorIs(err, expectedError)

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_StatusCode() {
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

	m := prepareSendgridMailV3(sender, recipients, nil, nil, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusServiceUnavailable}, error(nil))
	suite.Assert().Error(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_Success() {
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

	m := prepareSendgridMailV3(sender, recipients, nil, nil, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, nil, nil, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestSendEmail_MultipleEmails_Success() {
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

	m := prepareSendgridMailV3(sender, recipients, cc, bcc, subject, htmlContent)

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))
	suite.Assert().NoError(suite.emailSender.Send(ctx, sender, recipients, cc, bcc, subject, templatePath, data))

	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}

func (suite *SendgridGoTemplatesTestSuite) TestParseEmail() {
	recipients := []string{"test2@gazebosim.org", "test3@gazebosim.org"}

	expected := []*mail.Email{
		mail.NewEmail("", recipients[0]),
		mail.NewEmail("", recipients[1]),
	}

	result := parseSendgridEmails(recipients)

	suite.Assert().Equal(expected, result)
}

func TestSendgrid_EmailBuilder(t *testing.T) {
	suite.Run(t, new(SendgridEmailBuilderTestSuite))
}

type SendgridEmailBuilderTestSuite struct {
	suite.Suite
	builder sendgridEmailBuilder
}

func (suite *SendgridEmailBuilderTestSuite) SetupTest() {
	suite.builder = sendgridEmailBuilder{
		personalization: mail.NewPersonalization(),
		mail:            mail.NewV3Mail(),
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestSender() {
	const sender = "noreply@gazebosim.org"
	result := suite.builder.Sender(sender).personalization.From.Address
	suite.Assert().Equal(sender, result)
}

func (suite *SendgridEmailBuilderTestSuite) TestRecipients() {
	recipients := []string{"test1@gazebosim.org", "test2@gazebosim.org"}
	result := suite.builder.Recipients(recipients).personalization.To
	for _, to := range result {
		suite.Assert().Contains(recipients, to.Address)
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestCCs() {
	recipients := []string{"test1@gazebosim.org", "test2@gazebosim.org"}
	result := suite.builder.CC(recipients).personalization.CC
	for _, cc := range result {
		suite.Assert().Contains(recipients, cc.Address)
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestBBCs() {
	recipients := []string{"test1@gazebosim.org", "test2@gazebosim.org"}
	result := suite.builder.BCC(recipients).personalization.BCC
	for _, bcc := range result {
		suite.Assert().Contains(recipients, bcc.Address)
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestSubject() {
	const subject = "Welcome to Gazebo Sim"
	result := suite.builder.Subject(subject).personalization.Subject
	suite.Assert().Equal(subject, result)
}

func (suite *SendgridEmailBuilderTestSuite) TestMailContent() {
	const htmlContent = "<h1>Welcome to Gazebo Sim</h1>"
	const textContent = "Welcome to Gazebo Sim"

	contents := []string{htmlContent, textContent}

	result := suite.builder.Content("text/html", htmlContent).Content("text/plain", textContent).mail.Content

	for _, content := range result {
		suite.Assert().Contains(contents, content.Value)
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestTemplate() {
	const key = "template-id-123456789"
	data := struct {
		Test string `structs:"test"`
	}{
		Test: "test_data",
	}

	suite.builder = suite.builder.Template(key, data)

	suite.Assert().Equal(key, suite.builder.mail.TemplateID)
	suite.Assert().Contains(suite.builder.personalization.DynamicTemplateData, "test")
	suite.Assert().Equal("test_data", suite.builder.personalization.DynamicTemplateData["test"])
}

func (suite *SendgridEmailBuilderTestSuite) TestBuild() {
	const key = "template-id-123456789"
	data := struct {
		Test string `sendgrid:"test"`
	}{
		Test: "test_data",
	}

	m := suite.builder.
		Sender("noreply@gazebosim.org").
		Recipients([]string{"test@gazebosim.org"}).
		CC([]string{"cc@gazebosim.org"}).
		BCC([]string{"bcc@gazebosim.org"}).
		Subject("Test email subject").
		Template(key, data).
		Build()
	suite.Assert().NotNil(m)
}

func TestSendgrid_ContentInjector(t *testing.T) {
	suite.Run(t, new(SendgridContentInjectorTestSuite))
}

type SendgridContentInjectorTestSuite struct {
	suite.Suite
	builder sendgridEmailBuilder
}

func (suite *SendgridContentInjectorTestSuite) SetupTest() {
	suite.builder = sendgridEmailBuilder{
		personalization: mail.NewPersonalization(),
		mail:            mail.NewV3Mail(),
	}
}

func (suite *SendgridContentInjectorTestSuite) TestHTMLInjector() {
	var err error
	type emailData struct {
		Test string
	}

	suite.builder, err = injectHTMLContent(suite.builder, templatePath, emailData{Test: templatePath})
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(suite.builder.mail.Content)
	suite.Assert().Len(suite.builder.mail.Content, 1)
	suite.Assert().Contains(suite.builder.mail.Content[0].Value, "Open Robotics Team")
}

func (suite *SendgridContentInjectorTestSuite) TestHTMLInjector_WithError() {
	var err error
	suite.builder, err = injectHTMLContent(suite.builder, filepath.Join(os.TempDir(), "test.gohtml"), nil)
	suite.Assert().Error(err)
	suite.Assert().Empty(suite.builder.mail.Content)
}

func (suite *SendgridContentInjectorTestSuite) TestTemplateInjector() {
	var err error
	type emailData struct {
		Test string `structs:"test"`
	}

	templateID := "template-id-123456789"
	suite.builder, err = injectTemplateContent(suite.builder, templateID, emailData{Test: templateID})
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(suite.builder.personalization.DynamicTemplateData)
	suite.Assert().Len(suite.builder.personalization.DynamicTemplateData, 1)
	suite.Assert().Contains(suite.builder.personalization.DynamicTemplateData["test"], templateID)
}

type sendgridMock struct {
	mock.Mock
}

func (s *sendgridMock) SendWithContext(ctx context.Context, email *mail.SGMailV3) (*rest.Response, error) {
	args := s.Called(ctx, email)
	return args.Get(0).(*rest.Response), args.Error(1)
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
