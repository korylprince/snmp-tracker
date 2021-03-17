package snmp

import (
	"fmt"
	"net"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	snmpARPTablePrefix        = ".1.3.6.1.2.1.4.35.1.4"
	snmpPhysicalAddressTypeIP = "1"
)

type Arp struct {
	MacAddress string
	IPAddress  string
}

func getARPs(snmp *gosnmp.GoSNMP) ([]*Arp, error) {
	pdus, err := walkOIDs(snmp, []string{
		snmpARPTablePrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to walk for ARP table: %w", err)
	}

	arps := make([]*Arp, 0, len(pdus[snmpARPTablePrefix]))

	for _, pdu := range pdus[snmpARPTablePrefix] {
		oid := strings.TrimPrefix(pdu.Name, string(snmpARPTablePrefix)+".")
		split := strings.Split(oid, ".")
		if len(split) != 7 || split[1] != snmpPhysicalAddressTypeIP {
			return nil, fmt.Errorf("Unknown ARP table entry: %s", oid)
		}
		ip := strings.Join(split[3:], ".")
		if net.ParseIP(ip) == nil {
			return nil, fmt.Errorf("Unable to parse IP: %s", ip)
		}
		mac := net.HardwareAddr(pdu.Value.([]byte))
		if mac.String() == unknownMacAddress {
			continue
		}
		arps = append(arps, &Arp{MacAddress: mac.String(), IPAddress: ip})
	}

	return arps, nil
}
