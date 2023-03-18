package snmp

import (
	"fmt"
	"net"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	snmpLLDPSystemPrefix      = ".1.0.8802.1.1.2.1.4.1.1.9.0"
	snmpLLDPPortPrefix        = ".1.0.8802.1.1.2.1.4.1.1.8.0"
	snmpLLDPPortSubTypePrefix = ".1.0.8802.1.1.2.1.4.1.1.6.0"
	snmpLLDPMacAddressPrefix  = ".1.0.8802.1.1.2.1.4.1.1.7.0"

	snmpLLDPSubTypeMacAddress = 3
)

// LLDP is an LLDP record
type LLDP struct {
	LocalPort  *Port
	RemotePort *Port
}

func getLLDPs(snmp *gosnmp.GoSNMP, portTbl map[string]*Port) ([]*LLDP, error) {
	pdus, err := walkOIDs(snmp, []string{
		snmpLLDPSystemPrefix,
		snmpLLDPPortPrefix,
		snmpLLDPPortSubTypePrefix,
		snmpLLDPMacAddressPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to walk for LLDP table: %w", err)
	}

	cache := make(map[string]*LLDP)
	lldps := make([]*LLDP, 0, len(pdus[snmpLLDPSystemPrefix]))

	for _, pdu := range pdus[snmpLLDPSystemPrefix] {
		oid := strings.TrimPrefix(pdu.Name, string(snmpLLDPSystemPrefix))
		split := strings.Split(oid, ".")
		if len(split) != 3 {
			return nil, fmt.Errorf("Error parsing id: Expected split 3, got %d", len(split))
		}
		id := "." + split[1]
		//guard against empty duplicates
		rSysName := string(pdu.Value.([]byte))
		if rSysName != "" {
			l := &LLDP{LocalPort: portTbl[id], RemotePort: &Port{SystemName: rSysName}}
			cache[oid] = l
			lldps = append(lldps, l)
		}
	}
	for _, pdu := range pdus[snmpLLDPPortPrefix] {
		oid := strings.TrimPrefix(pdu.Name, string(snmpLLDPPortPrefix))
		portName := string(pdu.Value.([]byte))
		if cache[oid] != nil && portName != "" {
			cache[oid].RemotePort.Name = portName
		}
	}
	for _, pdu := range pdus[snmpLLDPPortSubTypePrefix] {
		oid := strings.TrimPrefix(pdu.Name, string(snmpLLDPPortSubTypePrefix))
		if pdu.Value.(int) == snmpLLDPSubTypeMacAddress && cache[oid] != nil {
			cache[oid].RemotePort.MacAddress = "valid"
		}
	}

	for _, pdu := range pdus[snmpLLDPMacAddressPrefix] {
		oid := strings.TrimPrefix(pdu.Name, string(snmpLLDPMacAddressPrefix))
		mac := net.HardwareAddr(pdu.Value.([]byte))
		if cache[oid] != nil && cache[oid].RemotePort.MacAddress == "valid" {
			cache[oid].RemotePort.MacAddress = mac.String()
		}
	}

	filtered := make([]*LLDP, 0, len(lldps))
	for _, l := range lldps {
		if l.RemotePort.Name != "" && l.RemotePort.MacAddress != "" && l.RemotePort.MacAddress != unknownMacAddress {
			filtered = append(filtered, l)
		}
	}

	return filtered, nil
}
