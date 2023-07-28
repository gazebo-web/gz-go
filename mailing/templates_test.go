package mailing

import (
	"context"
	"errors"
	"github.com/gazebo-web/gz-go/v8"
	"github.com/sendgrid/rest"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

func TestWithTemplates(t *testing.T) {
	suite.Run(t, new(TemplatesTestSuite))
}

type TemplatesTestSuite struct {
	suite.Suite
	sender     *templateWrapper
	client     sendgridMock
	baseSender Sender
}

func (suite *TemplatesTestSuite) SetupTest() {
	suite.client = sendgridMock{}
	suite.baseSender = NewSendgridEmailSender(&suite.client)
	suite.sender = NewTemplateWrapper(suite.baseSender, map[string]string{
		"test": "./testdata/template.gohtml",
	}).(*templateWrapper)
}

func (suite *TemplatesTestSuite) TearDownTest() {

}

func (suite *TemplatesTestSuite) TestSend_InvalidTemplate() {
	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}

	err := suite.sender.Send(context.Background(), "test@test.com", []string{"test@test.com"}, nil, nil, "Test", "notfound", data)
	suite.Assert().Error(err)

	suite.Assert().ErrorIs(err, ErrTemplateNotFound)
}

func (suite *TemplatesTestSuite) TestSend_Fail() {
	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}
	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3("test@test.com", []string{"test@test.com"}, nil, nil, "Test", htmlContent)

	ctx := context.Background()
	expectedError := errors.New("sendgrid failure")
	suite.client.On("SendWithContext", ctx, m).Return((*rest.Response)(nil), expectedError)

	err = suite.sender.Send(ctx, "test@test.com", []string{"test@test.com"}, nil, nil, "Test", "test", nil)
	suite.Assert().Error(err)
}

func (suite *TemplatesTestSuite) TestSend_Success() {
	type emailData struct {
		Test string
	}

	data := emailData{Test: "Hello there!"}
	htmlContent, err := gz.ParseHTMLTemplate(templatePath, data)
	suite.Require().NoError(err)

	m := prepareSendgridMailV3("test@test.com", []string{"test@test.com"}, nil, nil, "Test", htmlContent)

	ctx := context.Background()

	suite.client.On("SendWithContext", ctx, m).Return(&rest.Response{StatusCode: http.StatusOK}, error(nil))

	err = suite.sender.Send(ctx, "test@test.com", []string{"test@test.com"}, nil, nil, "Test", "test", data)
	suite.Assert().NoError(err)
	suite.client.AssertCalled(suite.T(), "SendWithContext", ctx, m)
}
