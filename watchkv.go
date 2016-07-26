package caddyconsul

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type servicesIota int

const (
	none servicesIota = iota
	directories
	subdomain
	both
)

type domain struct {
	Config   string
	Services servicesIota
}

func (s *caddyfile) WatchKV(reload bool) {

	opts := api.QueryOptions{
		WaitIndex: s.lastIndex,
		WaitTime:  5 * time.Minute,
	}
	fmt.Println("Watching for", s.lastIndex, "or better")
	pairs, _, err := kv.List("caddy/", &opts)
	if err != nil {
		// this should probably be logged
		return
	}

	// TODO actually make a new one, don't just keep using the old one
	if s.domains == nil {
		s.domains = make(map[string]*domain)
	}
	for _, k := range pairs {
		key := strings.TrimLeft(k.Key, "caddy/")
		if k.ModifyIndex > s.lastIndex {
			fmt.Println("index now", k.ModifyIndex)
			s.lastIndex = k.ModifyIndex
		}
		if key == "" {
			continue
		}
		fmt.Println(k.Key)
		keybits := strings.SplitN(key, "/", 2)
		if keybits[1] == "" {
			continue
		}
		if s.domains[keybits[0]] == nil {
			s.domains[keybits[0]] = &domain{
				Config:   "",
				Services: none,
			}
		}
		if keybits[1] == "config" {
			s.domains[keybits[0]].Config = keybits[1]
		} else if keybits[1] == "services" {
			if string(k.Value) == "directories" {
				s.domains[keybits[0]].Services = directories
			} else if string(k.Value) == "subdomain" {
				s.domains[keybits[0]].Services = subdomain
			} else if string(k.Value) == "both" {
				s.domains[keybits[0]].Services = both
			} else {
				s.domains[keybits[0]].Services = none
			}
		}
	}
	contents := ""
	for address, domain := range s.domains {
		contents += buildConfig(address, *domain)
	}
	s.contents = contents

	fmt.Println("Generated config:")
	fmt.Println(s.contents)
	if reload {
		reloadCaddy()
	}
}
