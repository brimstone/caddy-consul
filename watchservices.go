package caddyconsul

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type service struct {
	Name    string
	Address string
	Port    int
}

func (s *caddyfile) WatchServices(reload bool) {

	opts := api.QueryOptions{
		WaitIndex: s.lastService,
		WaitTime:  5 * time.Minute,
	}
	if !reload {
		opts.WaitTime = time.Second
	}
	fmt.Println("Watching for new service with index", s.lastService, "or better")
	// TODO
	services, meta, err := catalog.Services(&opts)
	if err != nil {
		// TODO should probably handle this better
		return
	}

	if meta.LastIndex > s.lastService {
		s.lastService = meta.LastIndex
	}

	myservices := make(map[string]map[string][]*service)
	for servicename := range services {
		//fmt.Println("Service:", servicename)
		// Get all instances for this service
		instances, _, _ := catalog.Service(servicename, "", nil)
		// TODO should probably check error
		for _, instance := range instances {
			for _, tag := range instance.ServiceTags {
				if !strings.HasPrefix(tag, "urlprefix-") {
					continue
				}
				cleantag := strings.TrimPrefix(tag, "urlprefix-")
				keybits := strings.SplitN(cleantag, "/", 2)
				if len(keybits) < 2 {
					// Our urlprefix isn't long enough, needs at least one forward slash
					continue
				}
				// Add the / back
				keybits[1] = "/" + keybits[1]
				myservice := &service{
					Name:    instance.ServiceName,
					Address: instance.Address,
					Port:    instance.ServicePort,
				}
				// TODO handle subdomains
				if instance.ServiceAddress != "" {
					myservice.Address = instance.ServiceAddress
				}
				if myservices[keybits[0]] == nil {
					myservices[keybits[0]] = make(map[string][]*service)
				}
				myservices[keybits[0]][keybits[1]] = append(myservices[keybits[0]][keybits[1]], myservice)
				if s.domains[keybits[0]] == nil {
					s.domains[keybits[0]] = &domain{
						Config: "",
					}
				}
			}
		}
	}

	s.services = myservices

	s.buildConfig()

	if reload {
		reloadCaddy()
	}
}
