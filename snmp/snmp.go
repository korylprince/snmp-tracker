//go:generate stringer -output=type_string.go -type=LinkStatusType -trimprefix LinkStatus
package snmp

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
	"github.com/korylprince/ipscan/resolve"
)

const snmpSystemName = ".1.3.6.1.2.1.1.5.0"

func getOIDs(snmp *gosnmp.GoSNMP, oids []string) (map[string]gosnmp.SnmpPDU, error) {
	sOIDs := make([]string, 0, len(oids))
	sOIDs = append(sOIDs, oids...)

	pkt, err := snmp.Get(sOIDs)
	if err != nil {
		return nil, fmt.Errorf("Failed to get OIDs %v: %w", oids, err)
	}
	if pkt.Error != gosnmp.NoError {
		return nil, fmt.Errorf("Received Error from agent: %d: %v", pkt.ErrorIndex, pkt.Error)
	}

	res := make(map[string]gosnmp.SnmpPDU)
	for _, pdu := range pkt.Variables {
		res[pdu.Name] = pdu
	}

	return res, nil
}

func walkOIDs(snmp *gosnmp.GoSNMP, oids []string) (map[string][]gosnmp.SnmpPDU, error) {
	pdus := make(map[string][]gosnmp.SnmpPDU)
	for _, oid := range oids {
		p, err := snmp.BulkWalkAll(oid)
		if err != nil {
			return nil, fmt.Errorf("Failed to BulkWalk OID %v: %w", oid, err)
		}
		pdus[oid] = p
	}

	return pdus, nil
}

// NetInfo is information from a device
type NetInfo struct {
	Ports        []*Port
	LLDPs        []*LLDP
	MacAddresses []*MacAddress
	Arps         []*Arp
	Resolves     []*Resolve
}

// System is a device
type System struct {
	Hostname          string `json:"hostname"`
	Port              uint16 `json:"port"`
	*ConnectionConfig `json:"connection"`
}

// Read retrieves information from network devices
func (s *System) Read(resolver *resolve.Service) (*NetInfo, error) {
	snmp := s.ConnectionConfig.New(s.Hostname, s.Port)

	if err := snmp.Connect(); err != nil {
		return nil, fmt.Errorf("Failed to open SNMP connection: %w", err)
	}
	defer snmp.Conn.Close()

	arps, err := getARPs(snmp)
	if err != nil {
		return nil, fmt.Errorf("Failed getting ARP info: %w", err)
	}

	resChan := getResolves(resolver, arps)

	pdusGet, err := getOIDs(snmp, []string{
		snmpSystemName,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get system name: %w", err)
	}
	sysName := ""
	if pdu, ok := pdusGet[snmpSystemName]; ok {
		sysName = string(pdu.Value.([]byte))
	}

	portTbl, err := getPortTable(snmp, sysName)
	if err != nil {
		return nil, fmt.Errorf("Failed getting port table: %w", err)
	}

	lldps, err := getLLDPs(snmp, portTbl)
	if err != nil {
		return nil, fmt.Errorf("Failed getting LLDP info: %w", err)
	}

	macs, err := getMacAddresses(snmp, portTbl)
	if err != nil {
		return nil, fmt.Errorf("Failed getting MAC Address info: %w", err)
	}

	ports := make([]*Port, 0, len(portTbl))
	for _, p := range portTbl {
		ports = append(ports, p)
	}

	return &NetInfo{
		Ports:        ports,
		MacAddresses: macs,
		Arps:         arps,
		LLDPs:        lldps,
		Resolves:     <-resChan,
	}, nil
}
