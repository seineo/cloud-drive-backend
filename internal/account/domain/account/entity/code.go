package entity

import (
	"common/eventbus"
	"common/eventbus/account"
	"fmt"
	"math/rand"
	"sync"
)

type VerificationCode struct {
	email  string
	code   string
	events []eventbus.Event
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

func (cf *CodeFactory) NewVerificationCode(email string) *VerificationCode {
	const digitBytes = "0123456789"
	b := make([]byte, cf.digits)
	cf.mu.Lock() // rand.Source is not thread safe
	defer cf.mu.Unlock()
	for i := range b {
		b[i] = digitBytes[cf.r.Intn(len(digitBytes))]
	}
	codeObj := &VerificationCode{
		email: email,
		code:  string(b),
	}
	// 领域事件：验证码已生成
	codeObj.AddEvent(account.NewCodeGeneratedEvent(email, codeObj.Get()))
	return codeObj
}

func (v *VerificationCode) Get() string {
	return v.code
}

func (v *VerificationCode) AddEvent(event eventbus.Event) {
	v.events = append(v.events, event)
}

func (v *VerificationCode) GetEvents() []eventbus.Event {
	return v.events
}

func (v *VerificationCode) ClearEvents() {
	v.events = []eventbus.Event{}
}
