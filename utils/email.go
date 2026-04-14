package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	// Message composition
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	// Authentication
	auth := smtp.PlainAuth("", from, password, host)

	// Sending email
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	return nil
}
