package logrus_mail_hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"
)

const (
	time_formate = "20060102 15:04:05"
)

// MailHook :不需要身份验证的邮件发送hook
type MailHook struct {
	AppName string
	c       *smtp.Client
}

//MailAuthHook :需要身份验证的邮件发送hook
type MailAuthHook struct {
	AppName  string
	Host     string
	Port     int
	From     *mail.Address
	To       *mail.Address
	UserName string
	PassWord string
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

func NewMailAuthHook(app string, host string, port int, from string, to string, username string, password string) (*MailAuthHook, error) {
	// 创建一个有超时时间的tcp网络连接 检测是否目标host服务器在监听端口
	conn, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// //校验邮箱发送者和接受者
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}
	receiver, err := mail.ParseAddress(to)
	if err != nil {
		return nil, err
	}

	return &MailAuthHook{
		AppName:  app,
		Host:     host,
		Port:     port,
		From:     sender,
		To:       receiver,
		UserName: username,
		PassWord: password}, nil

}

func (m *MailHook) Levels() []logrus.Level {
	//设置关联的日志级别组
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (m *MailAuthHook) Levels() []logrus.Level {
	//设置关联的日志级别组
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

	message := createMessage(entry, m.AppName)
	if _, err = message.WriteTo(writeCloser); err != nil {
		return err
	}

	return nil
}

func (m *MailAuthHook) Fire(entry *logrus.Entry) error {
	// 设置身份验证信息
	auth := smtp.PlainAuth("", m.UserName, m.PassWord, m.Host)

	message := createMessage(entry, m.AppName)
	// 设置 连接指定服务器 身份认证 发送方和接收方 并且进行邮件发送
	err := smtp.SendMail(
		m.Host+":"+strconv.Itoa(m.Port),
		auth,
		m.From.Address,
		[]string{m.To.Address},
		message.Bytes(),
	)
	if err != nil {
		return err
	}
	return nil
}

func createMessage(entry *logrus.Entry, appname string) *bytes.Buffer {
	body := entry.Time.Format(time_formate) + " - " + entry.Message
	title := appname + " - " + entry.Level.String()
	//格式化换行输出json
	field, _ := json.MarshalIndent(entry.Data, "", "\t")
	content := fmt.Sprintf("Subject:%s\r\n\r\n%s\r\n\r\n%s", title, body, field)
	message := bytes.NewBufferString(content)

	return message
}
