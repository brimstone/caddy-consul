package caddyconsul

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type service struct {
	Name      string
	Address   string
	Port      int
	Directory bool
}

func (s *caddyfile) WatchServices(reload bool) {

	opts := api.QueryOptions{
		WaitIndex: s.lastIndex,
		WaitTime:  5 * time.Minute,
	}
	fmt.Println("Watching for", s.lastIndex, "or better")
	// TODO
	services, meta, err := catalog.Services(&opts)
	if err != nil {
		// TODO should probably handle this better
		return
	}

	if meta.LastIndex > s.lastIndex {
		s.lastIndex = meta.LastIndex
	}

	myservices := make(map[string][]*service)
	for servicename, tags := range services {
		fmt.Println("Service:", servicename, "tags", tags)
		// loop over each tag
		for _, tag := range tags {
			// if the tag doesn't have our prefix, skip it
			if !strings.HasPrefix(tag, "urlprefix-") {
				continue
			}
			// register our service as having the
			instances, _, _ := catalog.Service(servicename, tag, nil)
			// TODO should probably check error
			for _, instance := range instances {
				myservice := &service{
					Name:      instance.ServiceName,
					Address:   instance.Address,
					Port:      instance.ServicePort,
					Directory: true,
				}
				// TODO handle subdomains
				if instance.ServiceAddress != "" {
					myservice.Address = instance.ServiceAddress
				}
				myservices[instance.ServiceID] = append(myservices[instance.ServiceID], myservice)
				fmt.Printf("%#v\n\n", instance)
			}
		}
	}

	s.services = myservices
	fmt.Printf("%#v\n", myservices)
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
