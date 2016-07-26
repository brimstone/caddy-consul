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
	Config string
}

func (s *caddyfile) WatchKV(reload bool) {

	opts := api.QueryOptions{
		WaitIndex: s.lastIndex,
		WaitTime:  5 * time.Minute,
	}
	fmt.Println("Watching for new KV with index", s.lastIndex, "or better")
	pairs, meta, err := kv.List("caddy/", &opts)
	if meta.LastIndex > s.lastIndex {
		s.lastIndex = meta.LastIndex
	}
	if err != nil {
		fmt.Println(err)
		// this should probably be logged
		return
	}

	// TODO actually make a new one, don't just keep using the old one
	if s.domains == nil {
		s.domains = make(map[string]*domain)
	}
	for _, k := range pairs {
		key := strings.TrimLeft(k.Key, "caddy/")
		if key == "" {
			continue
		}
		fmt.Println(k.Key)
		keybits := strings.SplitN(key, "/", 2)
		if s.domains[keybits[0]] == nil {
			s.domains[keybits[0]] = &domain{
				Config: "",
			}
		}
		if keybits[1] == "" {
			continue
		}
		if keybits[1] == "config" {
			s.domains[keybits[0]].Config = keybits[1]
		}
	}
	contents := ""
	for address, domain := range s.domains {
		contents += buildConfig(address, *domain, s.services)
	}
	s.contents = contents

	fmt.Println("Generated config:")
	fmt.Println(s.contents)
	if reload {
		reloadCaddy()
	}
}
