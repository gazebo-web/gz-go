package mailing

import "regexp"

// validateEmailAddress validates the given email is a valid email address.
func validateEmailAddress(email string) bool {
	exp := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return exp.MatchString(email)
}

// validateEmail validates that the given parameters for an email are valid.
func validateEmail(sender string, recipients []string, cc []string, bcc []string, data any) error {
	if len(recipients) == 0 {
		return ErrEmptyRecipientList
	}
	if len(sender) == 0 {
		return ErrEmptySender
	}
	for _, r := range recipients {
		if ok := validateEmailAddress(r); !ok {
			return ErrInvalidRecipient
		}
	}
	if ok := validateEmailAddress(sender); !ok {
		return ErrInvalidSender
	}
	for _, r := range cc {
		if ok := validateEmailAddress(r); !ok {
			return ErrInvalidRecipient
		}
	}
	for _, r := range bcc {
		if ok := validateEmailAddress(r); !ok {
			return ErrInvalidRecipient
		}
	}

	if data == nil {
		return ErrInvalidData
	}
	return nil
}
