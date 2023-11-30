package validation

import (
	"fmt"
	"net/mail"
	"regexp"
)

func CheckEmail(email string) error {
	_, mailErr := mail.ParseAddress(email)
	if mailErr != nil {
		return fmt.Errorf("email is not valid: %w", mailErr)
	}
	return nil
}

func CheckRegexMatch(regex string, input string) error {
	pattern := regexp.MustCompile(regex)
	if !pattern.MatchString(input) {
		return fmt.Errorf("string %s does not match the pattern %s", input, regex)
	}
	return nil
}
