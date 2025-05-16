package mail

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"gopkg.in/gomail.v2"
	"strconv"
)

type Service struct {
	mail gomail.Dialer
}

func NewMailService(config *config.Config) *Service {
	port, err := strconv.Atoi(config.Mail.Port)
	if err != nil {
		panic("Invalid mail port")
	}

	mail := gomail.Dialer{
		Host:     config.Mail.Host,
		Port:     port,
		Username: config.Mail.Username,
		Password: config.Mail.Password,
	}

	return &Service{mail: mail}
}

func (m *Service) SendMail(to string, subject string, message string) error {
	mail := gomail.NewMessage()
	mail.SetHeader("From", "ryu4w@gmail.com")
	mail.SetHeader("To", to)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", message)

	if err := m.mail.DialAndSend(mail); err != nil {
		return err
	}

	return nil
}
