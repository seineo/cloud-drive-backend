package service

import (
	"github.com/alexedwards/argon2id"
	"github.com/sirupsen/logrus"
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

func TestArgon2id(t *testing.T) {
	input := "pa$$word"
	hash, err := argon2id.CreateHash(input, argon2id.DefaultParams)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(logrus.Fields{
		"input": input,
		"hash":  hash,
	}).Info("compare")
	// ComparePasswordAndHash performs a constant-time comparison between a
	// plain-text password and Argon2id hash, using the parameters and salt
	// contained in the hash. It returns true if they match, otherwise it returns
	// false.
	match, err := argon2id.ComparePasswordAndHash(input, hash)
	if err != nil {
		log.Fatal(err)
	}
	t.Log("match is:", match)
}
