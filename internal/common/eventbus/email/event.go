package email

import (
	"common/eventbus"
	"encoding/json"
	"time"
)

type CodeEmailSent struct {
	eventbus.EventBase
	Email           string `json:"email"`
	Code            string `json:"code"`
	ConsumedEventID int64  `json:"consumedEventID"`
}

func (c CodeEmailSent) GetID() int64 {
	return c.EventID
}

func (c CodeEmailSent) GetName() string {
	return c.EventName
}

func (c CodeEmailSent) GetOccurTime() time.Time {
	return c.OccurTime
}

func (c CodeEmailSent) Marshall() ([]byte, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

func NewCodeEmailSentEvent(email string, code string, consumedEventID int64) eventbus.Event {
	eventBase := eventbus.NewEventBase("codeEmailSent")
	return CodeEmailSent{
		EventBase:       eventBase,
		Email:           email,
		Code:            code,
		ConsumedEventID: consumedEventID,
	}
}
