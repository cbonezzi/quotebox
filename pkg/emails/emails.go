package emails

import (
	"log"
	"net/smtp"
)

type Email struct {
    to string
    from string
    subject string
    content string
}

func New(to string, from string, subject string, content string) *Email {
    return &Email{
        to: to,
        from: from,
        subject: subject,
        content: content,
    }
}

func (e *Email) Send() {
	pass := "IFFodoLHBPwe2YxWOmWg"

	msg := "From: " + e.from + "\n" +
		"To: " + e.to + "\n" +
		"Subject: Hello there\n\n" +
		e.content

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", e.from, pass, "smtp.gmail.com"),
		e.from, []string{e.to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
