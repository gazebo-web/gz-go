package mailing

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/suite"
	"testing"
)

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
	builder := suite.builder.Sender(sender)
	suite.Assert().Equal(sender, builder.mail.From.Address)
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
