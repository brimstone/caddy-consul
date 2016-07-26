package caddyconsul

import (
	"os"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/mholt/caddy"
)

var consulClient *api.Client
var kv *api.KV
var catalog *api.Catalog

var consulGenerator *caddyfile

var initalized = false

type caddyfile struct {
	contents  string
	lastIndex uint64
	domains   map[string]*domain
	services  map[string][]*service
}

func (s *caddyfile) Body() []byte {
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

func init() {
	caddy.RegisterCaddyfileLoader("myloader", caddy.LoaderFunc(myLoader))
}

func reloadCaddy() {
	self, _ := os.FindProcess(os.Getpid())
	self.Signal(syscall.SIGUSR1)
}

func myLoader(serverType string) (caddy.Input, error) {
	if initalized {
		return consulGenerator, nil
	}
	consulAddress := os.Getenv("CONSUL")
	if consulAddress == "" {
		consulAddress = "127.0.0.1:8500"
	}
	var err error
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddress
	consulClient, err = api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	kv = consulClient.KV()
	catalog = consulClient.Catalog()

	consulGenerator = new(caddyfile)
	consulGenerator.WatchKV(false)
	consulGenerator.StartWatching()

	initalized = true
	return consulGenerator, nil
}

func buildConfig(address string, d domain, s map[string][]*service) string {
	ret := address + "\n"

	ret += d.Config + "\n"

	for servicename, _ := range s {
		ret += "#" + servicename + "\n"
	}
	return ret + "\n"
}
