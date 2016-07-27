package caddyconsul

import (
	"fmt"
	"time"
)

var consulGenerator *caddyfile

type caddyfile struct {
	contents    string
	lastKV      uint64
	lastService uint64
	domains     map[string]*domain
	services    map[string][]*service
}

func (s *caddyfile) Body() []byte {
	fmt.Println("Generated config:")
	fmt.Println(s.contents)
	return []byte(s.contents)
}

func (s *caddyfile) Path() string {
	return ""
}

func (s *caddyfile) ServerType() string {
	return "http"
}

func (s *caddyfile) StartWatching() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			s.WatchKV(true)
		}
	}()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			s.WatchServices(true)
		}
	}()
}
