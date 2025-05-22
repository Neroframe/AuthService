package gomail

import gomail "gopkg.in/mail.v2"

type Config struct {
	From         string
	Host         string
	Port         int
	SMTPUsername string
	SMTPPassword string
}

type Sender struct {
	from   string
	dialer *gomail.Dialer
}

func New(cfg Config) *Sender {
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.SMTPUsername, cfg.SMTPPassword)
	return &Sender{
		from:   cfg.From,
		dialer: d,
	}
}

func (s *Sender) Send(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	return s.dialer.DialAndSend(m)
}
