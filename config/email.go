package config

type EmailConfig struct {
	SMTPHost      string
	SMTPPort      int
	FromEmail     string
	FromEmailUser string
	FromEmailPsw  string
}
