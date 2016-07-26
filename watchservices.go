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
	fmt.Println("Watching for", s.lastService, "or better")
	// TODO
	services, meta, err := catalog.Services(&opts)
	if err != nil {
		// TODO should probably handle this better
		return
	}

	if meta.LastIndex > s.lastService {
		s.lastService = meta.LastIndex
	}

	myservices := make(map[string][]*service)
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
				myservice := &service{
					Name:    instance.ServiceName,
					Address: instance.Address,
					Port:    instance.ServicePort,
				}
				// TODO handle subdomains
				if instance.ServiceAddress != "" {
					myservice.Address = instance.ServiceAddress
				}
				myservices[cleantag] = append(myservices[cleantag], myservice)
				//fmt.Printf("%#v\n\n", instance)
			}
		}
	}

	s.services = myservices
	//fmt.Printf("%#v\n", myservices)
	contents := ""
	for address, domain := range s.domains {
		contents += buildConfig(address, *domain, s.services)
	}
	s.contents = contents

	if reload {
		reloadCaddy()
	}
}
