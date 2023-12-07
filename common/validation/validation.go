package validation

import (
	"CloudDrive/common/slugerror"
	"crypto/sha256"
	"encoding/base64"
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

// SHA256Hash concat strings and hash them using SHA256
func SHA256Hash(args ...string) string {
	h := sha256.New()
	var input string
	for _, str := range args {
		input = input + str
	}
	h.Write([]byte(input))
	// for fitting in url, we use base64 encoding
	// if we want to display it to users, we can use hex encoding
	// if store it as value in database, we should store raw bytes
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
