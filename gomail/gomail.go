package gomail

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/mail"
	"net/smtp"
	"time"
)

//
type Configuration struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

var Config = Configuration{
	Host:     "smtp.qq.com",
	Port:     "25",
	Username: "",
	Password: "",
	From:     "",
}

type GoMail struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Content string
}

func ParseMailAddr(address string) *mail.Address {
	addr, err := mail.ParseAddress(address)
	if err != nil {
		log.Fatalf("Parse mail address %s err: %s", address, err)
	}
	return addr
}

func (gm *GoMail) Send() error {
	to := make([]string, len(gm.To))
	for i := range gm.To {
		to[i] = ParseMailAddr(gm.To[i]).Address
	}

	if gm.From == "" {
		gm.From = Config.From
	}
	from := ParseMailAddr(gm.From).Address
	addr := fmt.Sprintf("%s:%s", Config.Host, Config.Port)
	auth := smtp.PlainAuth("", Config.Username, Config.Password, Config.Host)
	return smtp.SendMail(addr, auth, from, to, []byte(gm.String()))
}

func (gm *GoMail) String() string {
	var buf bytes.Buffer
	const crlf = "\r\n"

	write := func(what string, addrs []string) {
		if len(addrs) == 0 {
			return
		}
		for i := range addrs {
			if i == 0 {
				buf.WriteString(what)
			} else {
				buf.WriteString(", ")
			}
			buf.WriteString(ParseMailAddr(addrs[i]).String())
		}
		buf.WriteString(crlf)
	}
	getBoundary := func() string {
		h := md5.New()
		io.WriteString(h, fmt.Sprintf("%d", time.Now().Nanosecond()))
		return fmt.Sprintf("%x", h.Sum(nil))
	}

	from := ParseMailAddr(gm.From)
	if from.Address == "" {
		from = ParseMailAddr(Config.From)
	}
	fmt.Fprintf(&buf, "From: %s%s", from.String(), crlf)
	write("To: ", gm.To)
	write("Cc: ", gm.Cc)
	write("Bcc: ", gm.Bcc)
	boundary := getBoundary()
	fmt.Fprintf(&buf, "Date: %s%s", time.Now().UTC().Format(time.RFC822), crlf)
	fmt.Fprintf(&buf, "Subject: %s%s", gm.Subject, crlf)
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%s%s%s", boundary, crlf, crlf)
	fmt.Fprintf(&buf, "%s%s", "--"+boundary, crlf)
	fmt.Fprintf(&buf, "Content-Type: text/HTML; charset=UTF-8%s", crlf)
	fmt.Fprintf(&buf, "%s%s%s%s", crlf, gm.Content, crlf, crlf)
	fmt.Fprintf(&buf, "%s%s", "--"+boundary+"--", crlf)

	return buf.String()
}
