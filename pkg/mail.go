// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"bytes"
	"fmt"
	"github.com/toorop/go-dkim"
	"io"
	"net/mail"
	"net/smtp"
	"time"
)

type message struct {
	host     *host
	to, from string
	id       string

	content []byte
}

// Set the date and the Message-Id
func (m *message) setMeta(r io.Reader) error {
	now := time.Now()
	ms, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}

	// Date
	ms.Header["Date"] = []string{now.Format(time.RFC1123Z)}

	// Message-Id
	m.id = fmt.Sprintf("<%d@%s>", now.Unix(), m.host.name)
	ms.Header["Message-ID"] = []string{m.id}

	// Regenrate the message content.
	buff := &bytes.Buffer{}

	for n, v := range ms.Header {
		for _, v := range v {
			buff.WriteString(n)
			buff.WriteString(": ")
			buff.WriteString(v)
			buff.WriteString("\r\n")
		}
	}
	buff.WriteString("\r\n")

	if _, err = io.Copy(buff, ms.Body); err != nil {
		return err
	}

	m.content = buff.Bytes()

	return m.dkim()
}

// dkim create the DKIM signature.
func (m *message) dkim() error {
	return dkim.Sign(&m.content, m.host.dkimOption)
}

// Send the message to the smtp.Client.
func (m *message) send(c *smtp.Client) error {
	if err := c.Mail(m.from); err != nil {
		return fmt.Errorf("[ERROR] mail:%s on Client.Mail(): %v", m.id, err)
	}

	if err := c.Rcpt(m.to); err != nil {
		return fmt.Errorf("[ERROR] mail:%s on Client.Rcpt(): %v", m.id, err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("[ERROR] mail:%s on Client.Data(): %v", m.id, err)
	}
	defer w.Close()

	w.Write(m.content)

	return nil
}
