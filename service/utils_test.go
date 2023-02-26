package service

import (
	"testing"
	"time"
)

func TestGenerateCode(t *testing.T) {
	n := 6
	lastCode := ""
	for i := 0; i < 5; i++ {
		code := GenerateCode(n)
		t.Logf("current code is %s", code)
		if len(code) == 6 && code != lastCode {
			t.Logf("%d time: pass", i)
		} else {
			t.Errorf("%d time: failed, current code is %s, last code is %s", i, code, lastCode)
		}
		lastCode = code
	}

}

func TestSHA256Hash(t *testing.T) {
	email := "liyuewei2000@bupt.edu.cn"
	currentTime := time.Now().String()
	t.Logf("hash string is : %s", SHA256Hash(email, currentTime))
}
