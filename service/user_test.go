package service

import "testing"

type userInfo struct {
	name  string
	email string
}

func TestCheckUser(t *testing.T) {
	users := []userInfo{
		{"", "noreplay@gmail.test"},
		{"test3", "3@"},
		{"test1", "1@test.com"},
		{"test2", "2@test.com.cn"},
	}
	expected := []int{0, 1, 2, 2} // 0: name error, 1: email error, 2: both right
	for i, user := range users {
		err := CheckUser(user.name, user.email)
		var result int
		if err == UserNameError {
			result = 0
		} else if err == EmailFormatError {
			result = 1
		} else {
			result = 2
		}
		if result != expected[i] {
			t.Errorf("user: %v fails, expected: %d, actual: %d", user, expected[i], result)
		} else {
			t.Logf("user: %v is expected", user)
		}
	}
}

func TestSendCodeEmail(t *testing.T) {
	email := "531443089@qq.com"
	code, err := SendCodeEmail(email)
	if err != nil {
		t.Errorf("send email error: %s", err.Error())
	} else {
		t.Logf("send email successfully, code is %s", code)
	}
}
