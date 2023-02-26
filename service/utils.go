package service

import (
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
