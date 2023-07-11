package mailing

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSES_ReturnsErrWhenRecipientsIsEmpty(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(nil, "example@test.org", "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyRecipientList, err)
}

func TestSES_ReturnsErrWhenSenderIsEmpty(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send([]string{"recipient@test.org"}, "", "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptySender, err)
}

func TestSES_ReturnsErrWhenRecipientIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send([]string{"ThisIsNotAValidEmail"}, "example@test.org", "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRecipient, err)
}

func TestSES_ReturnsErrWhenSenderIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send([]string{"recipient@test.org"}, "InvalidSenderEmail", "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSender, err)
}

func TestSES_ReturnsErrWhenInvalidPath(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send([]string{"recipient@test.org"}, "example@test.org", "Some test", "test", nil)
	assert.Error(t, err)
}

func TestSES_ReturnsErrWhenDataIsNil(t *testing.T) {
	s := NewSimpleEmailServiceSender(&fakeSESSender{})
	err := s.Send([]string{"recipient@test.org"}, "example@test.org", "Some test", "testdata/template.gohtml", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidData, err)
}

func TestSES_SendingSuccess(t *testing.T) {
	fake := fakeSESSender{}
	s := NewSimpleEmailServiceSender(&fake)
	err := s.Send([]string{"recipient@test.org"}, "example@test.org", "Some test", "testdata/template.gohtml", struct {
		Test string
	}{
		Test: "Hello there!",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, fake.Called)
}

// fakeSESSender fakes the sesiface.SESAPI interface.
type fakeSESSender struct {
	returnError bool
	Called      int
	sesiface.SESAPI
}

// SendEmail mocks the SendEmail method from the sesiface.SESAPI.
func (s *fakeSESSender) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	s.Called++
	if s.returnError {
		return nil, errors.New("fake error")
	}
	return nil, nil
}
