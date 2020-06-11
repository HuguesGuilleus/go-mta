// BSD 3-Clause License
// Copyright (c) 2020, See AUTHORS file
// All rights reserved.

package mta

import (
	"bufio"
	"os"
	"regexp"
	"time"
)

type AuthTest func(login, password string) (ok bool)

// Say if the login and the passord is knwoed in the server.authFile.
func (s *server) auth(login, password string) (ok bool) {
	if time.Now().Sub(s.authUpdate) > time.Minute {
		s.loadAuth()
	}

	return s.authList[login] == password
}

// Update the auth map.
func (s *server) loadAuth() {
	// TODO: sync differente goruntime to minimise read.

	s.authUpdate = time.Now()
	if info, err := os.Stat(s.authFile); err != nil {
		s.l.Printf("[UPDATE AUTH ERROR] file '%s': %v", s.authFile, err)
		return
	} else if s.authModified.Equal(info.ModTime()) {
		return
	}

	s.l.Println("[UPDATE AUTH] from", s.authFile)
	f, err := os.Open(s.authFile)
	if err != nil {
		s.l.Printf("[UPDATE AUTH ERROR] file '%s': %v", s.authFile, err)
		return
	}
	defer f.Close()

	m := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := authComment.ReplaceAllString(scanner.Text(), "")
		if l == "" {
			continue
		}
		if !authLine.MatchString(l) {
			s.l.Printf("[UPDATE AUTH] wired line: %q in %s", l, s.authFile)
			continue
		}
		login := authLine.ReplaceAllString(l, "$1")
		password := authLine.ReplaceAllString(l, "$2")
		m[login] = password
	}
	s.authList = m
}

var (
	authComment = regexp.MustCompile(`\s*#.*$`)
	authLine    = regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s*$`)
)
