package caddyconsul

import (
	"fmt"
	"strconv"
	"time"
)

var consulGenerator *caddyfile

type caddyfile struct {
	contents    string
	lastKV      uint64
	lastService uint64
	domains     map[string]*domain
	services    map[string]map[string][]*service
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

func (s *caddyfile) buildConfig() {

	ret := ""
	for domainName := range s.domains {
		ret += domainName + " {\n"
		ret += s.domains[domainName].Config + "\n"
		for servicePath := range s.services[domainName] {
			ret += "	proxy " + servicePath
			for _, i := range s.services[domainName][servicePath] {
				ret += " " + i.Address + ":" + strconv.Itoa(i.Port)
			}
			ret += " {\n"
			ret += "		proxy_header X-Real-IP {remote}\n"
			ret += "		proxy_header X-Forwarded-Proto {scheme}\n"
			ret += "	}\n"
			ret += "\n"
		}
		ret += "}\n" // close domain
	}
	s.contents = ret
}
