package mailer

import (
	"bytes"
	"context"
	"html/template"
	"time"

	gomail "gopkg.in/gomail.v2"
)

type MailSender struct {
	Dialer *gomail.Dialer
	From   string
}

func NewGoMailer(port int, host, username, password, from string) *MailSender {
	d := gomail.NewDialer(host, port, username, password)
	return &MailSender{Dialer: d, From: from}
}

func (m *MailSender) SendMail(ctx context.Context, templatePath, to, subject string, data interface{}) error {

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	// Render template into a buffer
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	msg := gomail.NewMessage()

	msg.SetHeader("From", m.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	errCh := make(chan error, 1)

	go func() {
		errCh <- m.Dialer.DialAndSend(msg)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	case <-time.After(15 * time.Second):
		return context.DeadlineExceeded
	}

}
