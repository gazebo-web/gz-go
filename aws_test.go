package gz

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Tests for aws file

// TestSendEmal tests the SendEmail func
func TestSendEmail(t *testing.T) {

	sender := "sender@email.org"
	recipient := "recipient@email.org"
	subject := "gz-go AWS SES test"
	body := "Hello from gz-go!"

	err := SendEmail(sender, recipient, subject, body)
	assert.Error(t, err, "Should not be able to send email through AWS SES")
}
