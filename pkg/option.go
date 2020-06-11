// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"io"
	"log"
)

type Option struct {
	// The log output
	Out io.Writer
	// Listen address
	Addrs []string
	// All the hosts
	Hosts []HostOption
	// Hosts []HostOption
	Passord string

	AuthTest AuthTest
	// the file that contain login and password. Used if AuthTest is nil
	AuthFile string
}

func Listen(opt *Option) {
	s := server{
		l:        log.New(opt.Out, "", log.LstdFlags),
		hosts:    make(map[string]*host, len(opt.Hosts)),
		authTest: opt.AuthTest,
		authFile: opt.AuthFile,
	}

	if s.authTest == nil {
		s.authTest = s.auth
		if s.authFile == "" {
			log.Fatal("no Option.AuthFile")
		}
	}

	for _, opt := range opt.Hosts {
		h, err := newHost(&opt)
		if err != nil {
			log.Fatal(err)
		}
		s.hosts[opt.Name] = h
	}

	for _, a := range opt.Addrs {
		go s.listen(a)
	}
}
