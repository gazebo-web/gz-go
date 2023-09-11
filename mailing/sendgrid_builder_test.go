package mailing

import (
	"fmt"
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
	const sender = "Gazebo Web:noreply@gazebosim.org"
	from := suite.builder.Sender(sender).personalization.From

	suite.Assert().Equal(sender, fmt.Sprintf("%s:%s", from.Name, from.Address))
}

func (suite *SendgridEmailBuilderTestSuite) TestRecipients() {
	recipients := []string{"Test 1:test1@gazebosim.org", "Test 2:test2@gazebosim.org"}
	result := suite.builder.Recipients(recipients).personalization.To
	for _, to := range result {
		suite.Assert().Contains(recipients, fmt.Sprintf("%s:%s", to.Name, to.Address))
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestCCs() {
	recipients := []string{"Test 1:test1@gazebosim.org", "Test 2:test2@gazebosim.org"}
	result := suite.builder.CC(recipients).personalization.CC
	for _, cc := range result {
		suite.Assert().Contains(recipients, fmt.Sprintf("%s:%s", cc.Name, cc.Address))
	}
}

func (suite *SendgridEmailBuilderTestSuite) TestBBCs() {
	recipients := []string{"Test 1:test1@gazebosim.org", "Test 2:test2@gazebosim.org"}
	result := suite.builder.BCC(recipients).personalization.BCC
	for _, bcc := range result {
		suite.Assert().Contains(recipients, fmt.Sprintf("%s:%s", bcc.Name, bcc.Address))
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
		Recipients([]string{"Test Example:test@gazebosim.org"}).
		CC([]string{"Test CC:cc@gazebosim.org"}).
		BCC([]string{"Test BCC:bcc@gazebosim.org"}).
		Subject("Test email subject").
		Template(key, data).
		Build()
	suite.Assert().NotNil(m)
}
