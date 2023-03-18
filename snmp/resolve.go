package snmp

import (
	"net"
	"sync"

	"github.com/korylprince/ipscan/resolve"
)

// Resolve is a hostname resolution
type Resolve struct {
	IPAddress string
	Hostname  string
}

func getResolves(resolver *resolve.Service, arps []*Arp) chan []*Resolve {
	var resolves []*Resolve
	mu := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	wg.Add(len(arps))
	for _, arp := range arps {
		go func(ip string) {
			defer wg.Done()
			addr := net.ParseIP(ip)
			if addr == nil {
				return
			}
			host, err := resolver.LookupAddr(addr)
			if host == "" || err != nil {
				return
			}
			//remove the . at the end of the resolved named. Going for usability vs pendanticism here
			if host[len(host)-1] == '.' {
				host = host[:len(host)-1]
			}
			mu.Lock()
			resolves = append(resolves, &Resolve{IPAddress: ip, Hostname: host})
			mu.Unlock()
		}(arp.IPAddress)
	}

	c := make(chan []*Resolve)

	go func() {
		wg.Wait()
		c <- resolves
	}()

	return c
}
