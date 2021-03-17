package snmp

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

const (
	snmpMacTablePortPrefix = ".1.3.6.1.2.1.17.7.1.2.2.1.2"

	unknownMacAddress = "00:00:00:00:00:00"
)

type MacAddress struct {
	MacAddress string
	Port       *Port
	Vlan       int
}

func getMacAddresses(snmp *gosnmp.GoSNMP, portTbl map[string]*Port) ([]*MacAddress, error) {
	pdus, err := walkOIDs(snmp, []string{
		snmpMacTablePortPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to walk for MAC table: %w", err)
	}

	macs := make([]*MacAddress, 0, len(pdus[snmpMacTablePortPrefix]))

	for _, pdu := range pdus[snmpMacTablePortPrefix] {
		//returned OID is .vlan.d.d.d.d.d.d where d is decimal version of MAC address
		id := strings.TrimPrefix(pdu.Name, string(snmpMacTablePortPrefix))

		split := strings.Split(id, ".")
		if len(split) != 8 {
			return nil, fmt.Errorf("Error parsing id: Expected split 8, got %d", len(split))
		}
		vlan, err := strconv.Atoi(split[1])
		if err != nil {
			return nil, fmt.Errorf("Error parsing id: Couldn't parse VLAN: %w", err)
		}
		var mac net.HardwareAddr
		for _, s := range split[2:] {
			dec, err := strconv.ParseUint(s, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("Error parsing id: Couldn't parse Mac address: %w", err)
			}
			mac = append(mac, uint8(dec))
		}
		if mac.String() == unknownMacAddress {
			continue
		}

		macs = append(macs, &MacAddress{
			MacAddress: mac.String(),
			Port:       portTbl["."+strconv.Itoa(pdu.Value.(int))],
			Vlan:       vlan,
		})
	}

	return macs, nil
}
