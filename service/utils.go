package service

import (
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"time"
)

// GenerateCode Generate n-digit verification code
func GenerateCode(n int) string {
	const digitBytes = "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = digitBytes[rand.Intn(len(digitBytes))]
	}
	return string(b)
}

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
