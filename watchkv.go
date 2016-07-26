package caddyconsul

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type domain struct {
	Config string
}

func (s *caddyfile) WatchKV(reload bool) {

	opts := api.QueryOptions{
		WaitIndex: s.lastKV,
		WaitTime:  5 * time.Minute,
	}
	if !reload {
		opts.WaitTime = time.Second
	}
	fmt.Println("Watching for new KV with index", s.lastKV, "or better")
	pairs, meta, err := kv.List("caddy/", &opts)
	if meta.LastIndex > s.lastKV {
		s.lastKV = meta.LastIndex
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
			s.domains[keybits[0]].Config = string(k.Value)
		}
	}
	contents := ""
	for address, domain := range s.domains {
		contents += buildConfig(address, *domain, s.services)
	}
	if contents == "" {
		fmt.Println("Contents are empty. Perhaps KV isn't configured correctly?")
	}
	s.contents = contents

	if reload {
		reloadCaddy()
	}
}
