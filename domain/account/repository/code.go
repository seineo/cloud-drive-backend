package repository

type CodeRepository interface {
	SetCode(codeKey string, codeValue string) error
	GetCode(codeKey string) (string, error)
}
