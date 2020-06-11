// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package main

import (
	"./pkg"
	"github.com/HuguesGuilleus/go-logoutput"
	"gopkg.in/ini.v1"
	"log"
	"os"
)

func main() {
	// Get the config file
	configFile := "/etc/go-mta/config.ini"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config, err := ini.Load(configFile)
	if err != nil {
		log.Fatalf("[CONFIG INIT ERROR] %q: %v", configFile, err)
	}

	// The general configuration
	d := config.Section("")
	opt := mta.Option{
		Out:      logoutput.New(d.Key("out").MustString("/var/log/go-mta/")),
		Addrs:    d.Key("addrs").Strings(" "),
		AuthFile: d.Key("login").String(),
	}

	// Get the host config
	for _, s := range config.Sections() {
		if s.Name() == "DEFAULT" {
			continue
		}
		opt.Hosts = append(opt.Hosts, mta.HostOption{
			Name:         s.Name(),
			DkimKey:      s.Key("dkim_key").String(),
			DkimSelector: s.Key("dkim_selecctor").String(),
			Cert:         s.Key("crt").String(),
			Key:          s.Key("key").String(),
		})
	}

	// Listen and serve
	log.Fatal(mta.Listen(&opt))
}
