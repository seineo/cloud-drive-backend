package service

import (
	"bytes"
	"email/config"
	"email/domain/infrastructure"
	"fmt"
	"html/template"
)

type CodeEmailContent struct {
	Code        string
	ProjectName string
	ProjectURL  string
}

type EmailService interface {
	SendVerificationCode(email string, code string) error
}

type emailService struct {
	senderEmail map[string]string // 按任务对应发送方邮箱名
	emailSender infrastructure.EmailSender
	configs     *config.Config // 用于填充邮件模板的项目相关信息
}

func NewEmailService(senderEmail map[string]string, emailSender infrastructure.EmailSender, configs *config.Config) EmailService {
	return &emailService{senderEmail: senderEmail, emailSender: emailSender, configs: configs}
}

func (e *emailService) SendVerificationCode(email string, code string) error {
	// 读取asset中code email的模板 填充字段
	tmpl, err := template.ParseFiles("assets/code-email.html")
	if err != nil {
		return err
	}
	emailContent := CodeEmailContent{
		Code:        code,
		ProjectName: e.configs.ProjectName,
		ProjectURL:  e.configs.ProjectURL,
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, emailContent); err != nil {
		return err
	}
	// 选择code任务对应的发送方邮箱，发送邮件
	subject := fmt.Sprintf("%s-邮件认证", e.configs.ProjectName)
	if err := e.emailSender.SendEmail(e.senderEmail["code"], email, subject, buf.String()); err != nil {
		return err
	}
	return nil
}
