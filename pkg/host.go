// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"crypto/tls"
	"fmt"
	"github.com/toorop/go-dkim"
	"io/ioutil"
	"net"
	"net/smtp"
	"strings"
)

type host struct {
	name         string
	dkimOption   dkim.SigOptions
	certificates []tls.Certificate
}

// for the dev
func (h *host) Println(args ...interface{}) {
	fmt.Println(args...)
}

type HostOption struct {
	// The name of the host
	Name string
	// Dkim PEM key file
	DkimKey string
	// The text DNS selector
	DkimSelector string
	// The certificate and key for the TLS connexion
	Cert, Key string
}

func newHost(opt *HostOption) (*host, error) {
	if opt.Name == "" || opt.DkimSelector == "" {
		return nil, fmt.Errorf("HostOption Nmae or DkimSelector is empty")
	}

	h := &host{
		name:       opt.Name,
		dkimOption: dkim.NewSigOptions(),
	}
	h.dkimOption.Domain = opt.Name
	h.dkimOption.Selector = opt.DkimSelector
	h.dkimOption.Canonicalization = "relaxed/relaxed"
	h.dkimOption.Headers = []string{"from", "to", "date", "message-id", "subject"}

	if k, err := ioutil.ReadFile(opt.DkimKey); err != nil {
		return nil, fmt.Errorf("Erorr when read Option.DkimKey file: %s", err)
	} else {
		h.dkimOption.PrivateKey = k
	}

	if cert, err := tls.LoadX509KeyPair(opt.Cert, opt.Key); err != nil {
		return nil, fmt.Errorf("Erorr when load certificate: %s", err)
	} else {
		h.certificates = []tls.Certificate{cert}
	}

	return h, nil
}

/* CONNEXION */

// Get a valid connexion
func (h *host) connect(to string, m *message) {
	mxs, err := net.LookupMX(strings.SplitN(to, "@", 2)[1])
	if err != nil {
		h.Println(fmt.Errorf("Error on MX resolution: %v", err))
		return
	}

	for _, mx := range mxs {
		c, err := h.open(mx.Host[:len(mx.Host)-1])
		if err != nil {
			h.Println(err)
			continue
		}
		defer c.Close()

		if err := m.send(c); err != nil {
			h.Println(err)
			continue
		}

		h.Println(m.id, "correct delivery")
		return
	}

	h.Println("Mail " + m.id + " was not delivery")
}

// Open a connexion to the server.
func (h *host) open(serv string) (*smtp.Client, error) {
	conn, err := smtp.Dial(serv + ":smtp")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] on open connexion: %v", err)
	}

	if err := conn.Hello(h.name); err != nil {
		return nil, fmt.Errorf("[ERROR] on send hello: %v", err)
	}

	config := &tls.Config{
		ServerName:   serv,
		Certificates: h.certificates,
	}
	if err := conn.StartTLS(config); err != nil {
		return nil, fmt.Errorf("[ERROR] on StartTLS: %v", err)
	}

	return conn, nil
}
