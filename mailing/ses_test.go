package mailing

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSES_ReturnsErrWhenRecipientsIsEmpty(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(context.Background(), "example@test.org", nil, nil, nil, "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyRecipientList, err)
}

func TestSES_ReturnsErrWhenSenderIsEmpty(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(context.Background(), "", []string{"recipient@test.org"}, nil, nil, "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptySender, err)
}

func TestSES_ReturnsErrWhenRecipientIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(context.Background(), "example@test.org", []string{"ThisIsNotAValidEmail"}, nil, nil, "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRecipient, err)
}

func TestSES_ReturnsErrWhenSenderIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(context.Background(), "InvalidSenderEmail", []string{"recipient@test.org"}, nil, nil, "Some test", "test.template", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSender, err)
}

func TestSES_ReturnsErrWhenInvalidPath(t *testing.T) {
	s := NewSimpleEmailServiceSender(nil)

	err := s.Send(context.Background(), "example@test.org", []string{"recipient@test.org"}, nil, nil, "Some test", "test", nil)
	assert.Error(t, err)
}

func TestSES_ReturnsErrWhenDataIsNil(t *testing.T) {
	s := NewSimpleEmailServiceSender(&fakeSESSender{})
	err := s.Send(context.Background(), "example@test.org", []string{"recipient@test.org"}, nil, nil, "Some test", templatePath, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidData, err)
}

func TestSES_ReturnsErrWhenCCIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(&fakeSESSender{})
	err := s.Send(context.Background(), "example@test.org", []string{"recipient@test.org"}, []string{"test.org"}, nil, "Some test", templatePath, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRecipient, err)
}

func TestSES_ReturnsErrWhenBCCIsInvalid(t *testing.T) {
	s := NewSimpleEmailServiceSender(&fakeSESSender{})
	err := s.Send(context.Background(), "example@test.org", []string{"recipient@test.org"}, nil, []string{"test.org"}, "Some test", templatePath, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRecipient, err)
}

func TestSES_SendingSuccess(t *testing.T) {
	fake := fakeSESSender{}
	s := NewSimpleEmailServiceSender(&fake)
	err := s.Send(context.Background(), "example@test.org", []string{"recipient@test.org"}, nil, nil, "Some test", templatePath, struct {
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
