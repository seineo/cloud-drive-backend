package service

import (
	"testing"
)

func TestSendCodeEmail(t *testing.T) {
	email := "liyuewei2000@bupt.edu.cn"
	code, err := SendCodeEmail(email)
	if err != nil {
		t.Error("failed to send email")
	} else {
		t.Logf("successfully send email to %s, code is %s", email, code)
	}
}
