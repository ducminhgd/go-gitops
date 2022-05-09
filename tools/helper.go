package tools

import (
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/subosito/gotenv"
)

func init() {
	err := gotenv.Load()
	if err != nil {
		log.Println("Cannot load env")
	}
}

func CheckFatal(err error) {
	if err != nil {
		log.Fatalf("[Gitops] - Error:  %v", err)
	}
}

func Getenv(key, fallBack string) string {

	value := os.Getenv(key)
	if len(value) == 0 {
		return fallBack
	}
	return value
}

func SendEmail(to, cc, bcc, body, subject string) bool {
	port := Getenv("SMTP_PORT", "587")
	smtpServerHost := Getenv("SMTP_SERVER", "smtp.gmail.com")
	senderEmail := Getenv("SMTP_USERNAME", "")
	senderPassword := Getenv("SMTP_PASSWORD", "")

	if smtpServerHost == "" || senderEmail == "" || senderPassword == "" {
		log.Println("[Gitops] - Error: sender info is required")
		return false
	}

	headers := "From: " + senderEmail
	headers = headers + "\nTo: " + to + ";"
	rcpts := to
	if len(cc) > 0 {
		headers += "\nCc: " + cc + ";"
		rcpts += "," + cc
	}
	if len(bcc) > 0 {
		headers += "\nBcc: " + bcc + ";"
		rcpts += "," + bcc
	}
	headers = strings.ReplaceAll(headers, ",", ";")

	smtpServer := smtpServerHost + ":" + port
	mime := "Content-Type: text/html; charset=UTF-8"
	body = string(markdown.ToHTML([]byte(body), nil, nil))
	msg := []byte(headers + "\nSubject: " + subject + "\n" + mime + "\n" + body + "\n")
	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServerHost)
	err := smtp.SendMail(smtpServer, auth, senderEmail, strings.Split(rcpts, ","), msg)

	if err != nil {
		log.Printf("[Gitops] - Error:  %v", err)
		return false
	}
	return true
}

func GetFile(filePath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return file, err
	}
	return file, err
}
