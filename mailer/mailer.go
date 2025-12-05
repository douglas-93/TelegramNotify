package mailer

import (
	"html/template"

	"github.com/wneessen/go-mail"
)

type Mailer struct{}

func (m *Mailer) GetNewMail(from, to, subject string) (*mail.Msg, error) {
	message := mail.NewMsg()
	if err := message.From(from); err != nil {
		return nil, err
	}
	if err := message.To(to); err != nil {
		return nil, err
	}
	message.Subject(subject)

	return message, nil
}

func (m *Mailer) SetSimpleString(message *mail.Msg, value string) {
	message.SetBodyString(mail.TypeTextPlain, value)
}

func (m *Mailer) SetHTMLBody(message *mail.Msg, value string) {
	message.SetBodyHTMLTemplate(template.New("email").Parse(value))
}

func (m *Mailer) SendMail(env map[string]string, message *mail.Msg) error {
	client, err := mail.NewClient(env["SMTP_SERVER"], mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(env["SMTP_USER"]), mail.WithPassword(env["SMTP_PASSWORD"]))
	if err != nil {
		return err
	}
	if err = client.DialAndSend(message); err != nil {
		return err
	}
	return nil
}
