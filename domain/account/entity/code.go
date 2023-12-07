package entity

import (
	"fmt"
	"math/rand"
	"sync"
)

type VerificationCode struct {
	code string
}

type CodeFactory struct {
	digits uint
	r      *rand.Rand
	mu     sync.Mutex
}

func NewCodeFactory(digits uint, seed int64) (*CodeFactory, error) {
	if digits < 4 || digits > 8 {
		return nil, fmt.Errorf("the digits of the code are out of range: [4, 8]")
	}
	return &CodeFactory{digits: digits, r: rand.New(rand.NewSource(seed))}, nil
}

func (cf *CodeFactory) NewVerificationCode() *VerificationCode {
	const digitBytes = "0123456789"
	b := make([]byte, cf.digits)
	cf.mu.Lock() // rand.Source is not thread safe
	defer cf.mu.Unlock()
	for i := range b {
		b[i] = digitBytes[cf.r.Intn(len(digitBytes))]
	}
	return &VerificationCode{code: string(b)}
}

func (v *VerificationCode) Get() string {
	return v.code
}
