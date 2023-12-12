package account

import (
	"common/eventbus"
	"encoding/json"
	"time"
)

type CodeGenerated struct {
	eventbus.EventBase
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (c CodeGenerated) GetID() int64 {
	return c.EventID
}

func (c CodeGenerated) GetName() string {
	return c.EventName
}

func (c CodeGenerated) GetOccurTime() time.Time {
	return c.OccurTime
}

func (c CodeGenerated) Marshall() ([]byte, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func NewCodeGeneratedEvent(email string, code string) eventbus.Event {
	return CodeGenerated{
		EventBase: eventbus.NewEventBase("codeGenerated"),
		Email:     email,
		Code:      code,
	}
}
