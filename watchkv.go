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
	if err != nil {
		fmt.Println(err)
		// this should probably be logged
		return
	}
	if meta.LastIndex > s.lastKV {
		s.lastKV = meta.LastIndex
	}
	// If there's nothing, at least put our KV value so the user isn't lost
	if len(pairs) == 0 {
		kv.Put(&api.KVPair{Key: "caddy/"}, nil)
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
		if len(keybits) < 2 {
			continue
		}
		if keybits[1] == "config" {
			s.domains[keybits[0]].Config = string(k.Value)
		}
	}
	s.buildConfig()

	if reload {
		reloadCaddy()
	}
}
