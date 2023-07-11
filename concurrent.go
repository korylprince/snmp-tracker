package main

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/korylprince/ipscan/resolve"
	"github.com/korylprince/snmp-tracker/snmp"
)

func conWorker(wg *sync.WaitGroup, in <-chan *snmp.System, out chan<- *snmp.NetInfo, resolver *resolve.Service, debugPath string) {
	for sys := range in {
		info, err := sys.Read(resolver)
		if err != nil {
			log.Printf("WARNING: Unable to read system %s:%d: %v\n", sys.Hostname, sys.Port, err)
			continue
		}
		out <- info
		if debugPath != "" {
			if err := writeDebug(filepath.Join(debugPath, sys.Hostname+".json"), info); err != nil {
				log.Printf("WARNING: could not write snmp info for %s: %v", sys.Hostname, err)
			}
		}
	}
	wg.Done()
}

func conAgg(in <-chan *snmp.NetInfo, out chan<- *snmp.NetInfo) {
	info := &snmp.NetInfo{}
	for s := range in {
		info.Ports = append(info.Ports, s.Ports...)
		info.MacAddresses = append(info.MacAddresses, s.MacAddresses...)
		info.Arps = append(info.Arps, s.Arps...)
		info.LLDPs = append(info.LLDPs, s.LLDPs...)
		info.Resolves = append(info.Resolves, s.Resolves...)
	}
	out <- info
}

// GetInfo retrieves SNMP information concurrently
func GetInfo(resolver *resolve.Service, systems []*snmp.System, workers int, debugPath string) *snmp.NetInfo {
	wg := new(sync.WaitGroup)
	sysChan := make(chan *snmp.System)
	aggChan := make(chan *snmp.NetInfo)
	outChan := make(chan *snmp.NetInfo)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go conWorker(wg, sysChan, aggChan, resolver, debugPath)
	}
	go conAgg(aggChan, outChan)

	for _, s := range systems {
		sysChan <- s
	}
	close(sysChan)

	wg.Wait()
	close(aggChan)

	return <-outChan
}
