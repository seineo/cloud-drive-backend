package validation

import (
	"CloudDrive/common/slugerror"
	"fmt"
	"net/mail"
	"regexp"
)

func CheckEmail(email string) error {
	_, mailErr := mail.ParseAddress(email)
	if mailErr != nil {
		return slugerror.NewSlugError(slugerror.ErrInvalidInput, mailErr.Error(), "email is invalid")
	}
	return nil
}

func CheckRegexMatch(regex string, input string) error {
	pattern := regexp.MustCompile(regex)
	if !pattern.MatchString(input) {
		return slugerror.NewSlugError(slugerror.ErrInvalidInput, "pattern not match",
			fmt.Sprintf("input %s does not match pattern %s", input, pattern))
	}
	return nil
}
