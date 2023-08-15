package mailing

import (
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

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
