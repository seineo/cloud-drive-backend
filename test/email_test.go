package test

import (
	"CloudDrive/service"
	"testing"
)

func TestSendCodeEmail(t *testing.T) {
	email := "liyuewei2000@bupt.edu.cn"
	code, err := service.SendCodeEmail(email)
	if err != nil {
		t.Error("failed to send email")
	} else {
		t.Logf("successfully send email to %s, code is %s", email, code)
	}
}
