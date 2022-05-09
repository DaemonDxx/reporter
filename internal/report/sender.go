package report

import (
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	log "github.com/sirupsen/logrus"
	"net/smtp"
)

const filename = "Отчет о температурном факторе.xlsm"

type Sender interface {
	Send(message string, report *Report) error
}

type MailSender struct {
	config      *MailConfig
	subscribers *[]string
	host        *string
	auth        *smtp.Auth
	tls         *tls.Config
	log         *log.Entry
}

func NewMailSender(subscribers *[]string, config *MailConfig) Sender {
	host := fmt.Sprintf("%s:%s", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Login, config.Password, config.Host)
	tls := tls.Config{
		ServerName: config.Host,
	}
	return &MailSender{
		config:      config,
		subscribers: subscribers,
		host:        &host,
		auth:        &auth,
		tls:         &tls,
		log:         log.New().WithField("module", "sender"),
	}
}

func (s *MailSender) Send(message string, report *Report) error {
	e := email.NewEmail()
	e.From = s.config.Login + "@yandex.ru"
	e.To = *s.subscribers
	e.Subject = "Температурный фактор"
	e.HTML = []byte(message)
	_, err := e.Attach(report.Data, filename, "application/vnd.ms-excel.sheet.macroEnabled.12")
	if err != nil {
		s.log.Errorf("Не удалось прикрепить файл к письму: %e", err)
		return err
	}
	err = e.SendWithTLS(*s.host, *s.auth, s.tls)
	if err != nil {
		s.log.Errorf("Ошибка отправки сообщения: %e", err)
	}
	return err
}
