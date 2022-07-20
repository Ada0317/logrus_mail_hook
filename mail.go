package logrus_mail_hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/mail"
	"net/smtp"
	"strconv"
)

const (
	time_formate = "20060102 15:04:05"
)

type MailHook struct {
	AppName string
	c       *smtp.Client
}

func NewMailHook(app string, host string, port int, from string, to string) (*MailHook, error) {
	//连接远程smtp服务器
	c, err := smtp.Dial(host + ":" + strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	defer c.Close()

	//校验邮箱发送者和接受者
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}
	receiver, err := mail.ParseAddress(to)
	if err != nil {
		return nil, err
	}

	//设置client的发送者和接收者
	if err = c.Mail(sender.Address); err != nil {
		return nil, err
	}
	if err = c.Rcpt(receiver.Address); err != nil {
		return nil, err
	}

	return &MailHook{
		AppName: app,
		c:       c,
	}, nil
}

func (m *MailHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (m *MailHook) Fire(entry *logrus.Entry) error {
	writeCloser, err := m.c.Data()
	if err != nil {
		return err
	}
	defer writeCloser.Close()
	message := CreateMessage(entry, m.AppName)
	if _, err = message.WriteTo(writeCloser); err != nil {
		return err
	}
	return nil
}

func CreateMessage(entry *logrus.Entry, appname string) *bytes.Buffer {
	body := entry.Time.Format(time_formate) + " - " + entry.Message
	title := appname + " - " + entry.Level.String()
	//格式化换行输出json
	field, _ := json.MarshalIndent(entry.Data, "", "\t")
	content := fmt.Sprintf("Subject:%s\r\n\r\n%s\r\n\r\n%s", title, body, field)
	message := bytes.NewBufferString(content)

	return message
}
