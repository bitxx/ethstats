package emailutil

import (
	"crypto/tls"
	"errors"
	"ethstats/server/config"
	"gopkg.in/gomail.v2"
	"strings"
	"sync"
)

var (
	wlock sync.Mutex
)

func SendEmailDefault(subject, content string) error {
	return SendEmail(config.EmailConfig.ToEmail, config.EmailConfig.SubjectPrefix+subject, content,
		config.EmailConfig.Username, config.EmailConfig.FromEmail, config.EmailConfig.Host,
		config.EmailConfig.Password, config.EmailConfig.ContentType, config.EmailConfig.Port)
}

// SendEmail send to notification
func SendEmail(toEmailList, subject, content string, username, fromEmail, host, password, contentType string, port int) error {
	if len(toEmailList) <= 1 || subject == "" || content == "" || username == "" || fromEmail == "" || host == "" || password == "" || contentType == "" || port <= 0 {
		return errors.New("param init error,not send email")
	}
	wlock.Lock()
	defer wlock.Unlock()

	emailSlips := strings.Split(toEmailList, ",")
	var emails []string
	for _, v := range emailSlips {
		emails = append(emails, v)
	}

	// smtp 配置
	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send emails using d.
	toEmails := emails

	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", toEmails...)

	// 邮件标题
	m.SetHeader("Subject", subject)

	m.SetBody(contentType, content)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
