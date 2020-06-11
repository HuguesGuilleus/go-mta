// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"crypto/tls"
	"fmt"
	"github.com/toorop/go-dkim"
	"io/ioutil"
)

type host struct {
	name        string
	dkimOption  dkim.SigOptions
	certificate []tls.Certificate
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
	h.dkimOption.Headers = []string{"from", "to", "date"}

	if k, err := ioutil.ReadFile(opt.DkimKey); err != nil {
		return nil, fmt.Errorf("Erorr when read Option.DkimKey file: %s", err)
	} else {
		h.dkimOption.PrivateKey = k
	}

	if cert, err := tls.LoadX509KeyPair(opt.Cert, opt.Key); err != nil {
		return nil, fmt.Errorf("Erorr when load certificate: %s", err)
	} else {
		h.certificate = []tls.Certificate{cert}
	}

	return h, nil
}
