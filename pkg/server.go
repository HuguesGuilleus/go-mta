// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

type server struct {
	l     *log.Logger
	hosts map[string]*host
	// Auth function s.auth() or a custom function from Option
	authTest AuthTest
	// The list of login and passord used by s.auth()
	authList map[string]string
	// The file that contain login and passord
	authFile string
	// The modified of the file.
	authModified time.Time
	// The last check of the authFile
	authUpdate time.Time
}

func (s *server) listen(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.l.Println("Error when listen:", addr, err)
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error when open an connexion:", err)
			continue
		}
		go s.newConn(conn)
	}
}

// Get the mail from this connexion and close it.
// TODO: make log ...
// TODO: check error for every ReadLine() and PrintfLine()
func (s *server) newConn(src net.Conn) {
	log.Println("[NEW CONN]", src.RemoteAddr())
	c := textproto.NewConn(src)
	defer c.Close()

	c.PrintfLine("220 SMTP Ready")
	c.ReadLine()
	c.PrintfLine("250-smtp.xxxx.xxxx")
	c.PrintfLine("250-PIPELINING")
	c.PrintfLine("250-AUTH PLAIN")
	c.PrintfLine("250 8BITMIME")

	// Auth
	l, _ := c.ReadLine()
	if !strings.HasPrefix(l, "AUTH PLAIN ") {
		c.PrintfLine("503 'AUTH PLAIN' expected")
		return
	}
	login, password, err := getAuth(l)
	if err != nil {
		c.PrintfLine("500 error when get plain auth")
		return
	}
	if !s.authTest(login, password) {
		c.PrintfLine("530 Auth fail")
		return
	}
	c.PrintfLine("235 2.7.0 Authentication successful")

	c.ReadLine()
	c.PrintfLine("250 Sender ok")

	// Get the Destination
	l, _ = c.ReadLine()
	to := regexp.MustCompile(`.*<(.*)>`).ReplaceAllString(l, "$1")
	log.Printf("to: %#+v\n", to)
	c.PrintfLine("250 Recipient ok.")

	// Data
	c.ReadLine()
	c.PrintfLine("354 Enter mail, end with \".\" on a line by itself")
	buff := bytes.Buffer{}
	for l, _ := c.ReadLine(); l != "."; l, _ = c.ReadLine() {
		buff.WriteString(l)
		buff.WriteString("\r\n")
	}

	go s.newMessage(login, to, &buff)

	c.PrintfLine("250 Ok")
	c.ReadLine()
	c.PrintfLine("221 Closing connection")
}

// Return the login, the password and an error of decoding.
func getAuth(line string) (string, string, error) {
	data, err := base64.StdEncoding.DecodeString(
		strings.TrimPrefix(line, "AUTH PLAIN "))
	if err != nil {
		return "", "", err
	}

	s := strings.SplitN(string(data), "\x00", 3)
	if len(s) != 3 {
		return "", "", fmt.Errorf("Need \\0+login\\0+password")
	}

	return s[1], s[2], nil
}

func (s *server) newMessage(login, to string, r io.Reader) {
	fmt.Println("newMessage()")
	m := message{
		host: s.getHost(login),
		to:   to,
		from: login,
	}

	if m.host == nil {
		return
	}

	fmt.Println("on s'occupe des méta données")

	if err := m.setMeta(r); err != nil {
		fmt.Println("m.setMeta() error:", err)
		return
	}
	fmt.Println("Le message est prêt pour l'envoie")

	m.host.connect(m.to, &m)
}

func (s *server) getHost(login string) *host {
	n := regexp.MustCompile(`\w+@`).ReplaceAllString(login, "")
	log.Printf("n: %#+v\n", n) //////////////
	h := s.hosts[n]
	log.Printf("h: %#+v\n", h) //////////////
	if h == nil {
		s.l.Println("Unknwon host for", login)
	}
	return h
}
