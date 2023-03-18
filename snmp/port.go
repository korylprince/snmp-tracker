//go:generate stringer -output=type_string.go -type=LinkStatusType
package snmp

import (
	"fmt"
	"net"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	snmpPortMacAddressPrefix  = ".1.3.6.1.2.1.2.2.1.6"
	snmpPortNamePrefix        = ".1.3.6.1.2.1.2.2.1.2"
	snmpPortDescriptionPrefix = ".1.3.6.1.2.1.31.1.1.1.18"
	snmpPortLinkStatusPrefix  = ".1.3.6.1.2.1.2.2.1.8"
	snmpPortSpeedPrefix       = ".1.3.6.1.2.1.31.1.1.1.15"
)

// LinkStatusType is type of link statuses
type LinkStatusType int

// link statuses
const (
	LinkStatusUp             LinkStatusType = 1
	LinkStatusDown           LinkStatusType = 2
	LinkStatusTesting        LinkStatusType = 3
	LinkStatusUnknown        LinkStatusType = 4
	LinkStatusDormant        LinkStatusType = 5
	LinkStatusNotPresent     LinkStatusType = 6
	LinkStatusLowerLayerDown LinkStatusType = 7
)

// Port is a switch port
type Port struct {
	SystemName  string
	MacAddress  string
	Name        string
	Description string
	LinkStatus  LinkStatusType
	Speed       uint
}

func getPortTable(snmp *gosnmp.GoSNMP, sysName string) (map[string]*Port, error) {
	pdus, err := walkOIDs(snmp, []string{
		snmpPortMacAddressPrefix,
		snmpPortNamePrefix,
		snmpPortDescriptionPrefix,
		snmpPortLinkStatusPrefix,
		snmpPortSpeedPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to walk for ports: %w", err)
	}

	tbl := make(map[string]*Port)

	for _, pdu := range pdus[snmpPortMacAddressPrefix] {
		id := strings.TrimPrefix(pdu.Name, string(snmpPortMacAddressPrefix))
		mac := net.HardwareAddr(pdu.Value.([]byte))
		if mac.String() != unknownMacAddress {
			tbl[id] = &Port{SystemName: sysName, MacAddress: mac.String()}
		}
	}
	for _, pdu := range pdus[snmpPortNamePrefix] {
		id := strings.TrimPrefix(pdu.Name, string(snmpPortNamePrefix))
		if port, ok := tbl[id]; ok {
			port.Name = string((pdu.Value).([]byte))
		}
	}
	for _, pdu := range pdus[snmpPortDescriptionPrefix] {
		id := strings.TrimPrefix(pdu.Name, string(snmpPortDescriptionPrefix))
		if port, ok := tbl[id]; ok {
			port.Description = string(pdu.Value.([]byte))
		}
	}
	for _, pdu := range pdus[snmpPortLinkStatusPrefix] {
		id := strings.TrimPrefix(pdu.Name, string(snmpPortLinkStatusPrefix))
		if port, ok := tbl[id]; ok {
			port.LinkStatus = LinkStatusType(pdu.Value.(int))
		}
	}
	for _, pdu := range pdus[snmpPortSpeedPrefix] {
		id := strings.TrimPrefix(pdu.Name, string(snmpPortSpeedPrefix))
		if port, ok := tbl[id]; ok {
			port.Speed = pdu.Value.(uint)
		}
	}

	return tbl, nil
}
