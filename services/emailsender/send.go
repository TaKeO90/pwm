package emailsender

import (
	"net/smtp"
	"os"
)

//EmailInfo struct holds email info
type EmailInfo struct {
	msg        []byte
	recipients []string
}

var (
	FROM      string = "pwm.noreply"
	password  string = os.Getenv("MAILPW")
	emailaddr string = os.Getenv("MAILADDR")
	hostname  string = "smtp.gmail.com"
	port      string = ":587"
)

//EmailAuth email authentication
func emailAuth() smtp.Auth {
	auth := smtp.PlainAuth("", emailaddr, password, hostname)
	return auth
}

func NewEmail(msg, email string) *EmailInfo {
	return &EmailInfo{[]byte(msg), []string{email}}
}

//SendCode sending code that we generate to the user
func (send *EmailInfo) SendCode() (bool, error) {
	auth := emailAuth()
	err := smtp.SendMail(hostname+port, auth, FROM, send.recipients, send.msg)
	if err != nil {
		return false, err
	}
	return true, nil
}
