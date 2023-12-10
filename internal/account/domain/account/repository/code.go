package repository

import (
	"time"
)

type CodeRepository interface {
	SetCode(codeKey string, code string, expiration time.Duration) error
	GetCode(codeKey string) (string, error)
}
