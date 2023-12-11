package infrastructure

type EmailSender interface {
	SendEmail(from string, to string, subject string, body string) error
}
