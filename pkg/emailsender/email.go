package emailsender

import (
	"fmt"
	"net/smtp"

	"github.com/aeilang/urlshortener/config"
	mail "github.com/jordan-wright/email"
)

type EmailSend struct {
	addr    string
	myMail  string
	subject string
	auth    smtp.Auth
}

func NewEmailSend(cfg config.EmailConfig) (*EmailSend, error) {
	emailSend := &EmailSend{
		addr:    fmt.Sprintf("%s:%s", cfg.HostAddress, cfg.HostPort),
		auth:    smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.HostAddress),
		myMail:  cfg.Username,
		subject: cfg.Subject,
	}
	if err := emailSend.Send(cfg.TestMail, "test email"); err != nil {
		return nil, err
	}

	return emailSend, nil
}

func (e *EmailSend) Send(email, emailCode string) error {
	instance := mail.NewEmail()
	instance.From = e.myMail
	instance.To = []string{email}
	instance.Subject = e.subject
	instance.Text = []byte(fmt.Sprintf("你的验证码为： %s", emailCode))

	return instance.Send(e.addr, e.auth)
}
