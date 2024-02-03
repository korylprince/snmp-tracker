package main

import (
	"fmt"
	"time"

	"github.com/korylprince/snmp-tracker/snmp"
)

// Upsert is a SQL upsert clause
type Upsert struct {
	Constraint    string   `json:"constraint"`
	UpdateColumns []string `json:"update_columns"`
}

var hostnameOnConflict = &Upsert{Constraint: "unique_hostname", UpdateColumns: []string{"hostname"}}
var systemOnConflict = &Upsert{Constraint: "unique_system_name", UpdateColumns: []string{"name"}}
var systemOnConflictHostname = &Upsert{Constraint: "unique_system_name", UpdateColumns: []string{"name", "hostname_id"}}
var macAddressOnConflict = &Upsert{Constraint: "unique_mac_address", UpdateColumns: []string{"mac_address"}}
var portOnConflict = &Upsert{Constraint: "unique_port_system_name", UpdateColumns: []string{"system_id", "name"}}
var portOnConflictMacAddressDescription = &Upsert{Constraint: "unique_port_system_name", UpdateColumns: []string{"system_id", "name", "mac_address_id", "description"}}
var lldpOnConflict = &Upsert{Constraint: "unique_lldp", UpdateColumns: []string{"local_port_id", "remote_port_id"}}
var ipAddressOnConflict = &Upsert{Constraint: "unique_ip_address", UpdateColumns: []string{"ip_address"}}
var arpOnConflict = &Upsert{Constraint: "unique_arp", UpdateColumns: []string{"mac_address_id", "ip_address_id"}}
var resolveOnConflict = &Upsert{Constraint: "unique_resolve", UpdateColumns: []string{"ip_address_id", "hostname_id"}}

// Hostname is a device hostname
type Hostname struct {
	Hostname string `json:"hostname"`
}

// HostnamePointer is a pointer to a Hostname
type HostnamePointer struct {
	Data       *Hostname `json:"data"`
	OnConflict *Upsert   `json:"on_conflict"`
}

// System is a device
type System struct {
	Name     string           `json:"name"`
	Hostname *HostnamePointer `json:"hostname,omitempty"`
}

// SystemPointer is a pointer to a System
type SystemPointer struct {
	Data       *System `json:"data"`
	OnConflict *Upsert `json:"on_conflict"`
}

// MacAddress is a MAC address
type MacAddress struct {
	MacAddress string `json:"mac_address"`
}

// MacAddressPointer is a pointer to a MacAddress
type MacAddressPointer struct {
	Data       *MacAddress `json:"data"`
	OnConflict *Upsert     `json:"on_conflict"`
}

// Port is a switch port
type Port struct {
	System      *SystemPointer     `json:"system"`
	MacAddress  *MacAddressPointer `json:"mac_address"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
}

// PortPointer is a pointer to a port
type PortPointer struct {
	Data       *Port   `json:"data"`
	OnConflict *Upsert `json:"on_conflict"`
}

// PortJournal is a journal of ports
type PortJournal struct {
	Port   *PortPointer `json:"port"`
	Time   *time.Time   `json:"time"`
	Status string       `json:"status"`
	Speed  int          `json:"speed"`
}

// LLDP is an LLDP record
type LLDP struct {
	LocalPort  *PortPointer `json:"local_port"`
	RemotePort *PortPointer `json:"remote_port"`
}

// LLDPPointer is a pointer to an LLDP
type LLDPPointer struct {
	Data       *LLDP   `json:"data"`
	OnConflict *Upsert `json:"on_conflict"`
}

// LLDPJournal is a journal or LLDP records
type LLDPJournal struct {
	LLDP *LLDPPointer `json:"lldp"`
	Time *time.Time   `json:"time"`
}

// MacAddressJournal is a journal of MAC addresses
type MacAddressJournal struct {
	MacAddress *MacAddressPointer `json:"mac_address"`
	Port       *PortPointer       `json:"port"`
	Time       *time.Time         `json:"time"`
	Vlan       int                `json:"vlan"`
}

// IPAddress is an IP address
type IPAddress struct {
	IPAddress string `json:"ip_address"`
}

// IPAddressPointer is a pointer to an IPAddress
type IPAddressPointer struct {
	Data       *IPAddress `json:"data"`
	OnConflict *Upsert    `json:"on_conflict"`
}

// Arp is an ARP record
type Arp struct {
	MacAddress *MacAddressPointer `json:"mac_address"`
	IPAddress  *IPAddressPointer  `json:"ip_address"`
}

// ArpPointer is a pointer to an ARP
type ArpPointer struct {
	Data       *Arp    `json:"data"`
	OnConflict *Upsert `json:"on_conflict"`
}

// ArpJournal is a journal of ARP records
type ArpJournal struct {
	Arp  *ArpPointer `json:"arp"`
	Time *time.Time  `json:"time"`
}

// Resolve is a hostname resolution record
type Resolve struct {
	IPAddress *IPAddressPointer `json:"ip_address"`
	Hostname  *HostnamePointer  `json:"hostname"`
}

// ResolvePointer is a pointer to a Resolve
type ResolvePointer struct {
	Data       *Resolve `json:"data"`
	OnConflict *Upsert  `json:"on_conflict"`
}

// ResolveJournal is a journal of hostname resolution records
type ResolveJournal struct {
	Resolve *ResolvePointer `json:"resolve"`
	Time    *time.Time      `json:"time"`
}

// Journal is a journal of records
type Journal struct {
	Ports        []*PortJournal
	LLDPs        []*LLDPJournal
	MacAddresses []*MacAddressJournal
	Arps         []*ArpJournal
	Resolves     []*ResolveJournal
}

func portKey(p *snmp.Port) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", p.SystemName, p.Name)
}

// Translate translates SNMP info to a Journal
func Translate(i *snmp.NetInfo) *Journal {
	j := new(Journal)
	t := time.Now().UTC()

	sysCache := make(map[string]*SystemPointer)
	portCache := make(map[string]*PortPointer)

	for _, p := range i.Ports {
		var sp *SystemPointer
		if s, ok := sysCache[p.MacAddress]; ok {
			sp = s
		} else {
			sp = &SystemPointer{Data: &System{Name: p.SystemName}, OnConflict: systemOnConflict}
			sysCache[p.MacAddress] = sp
		}
		mp := &MacAddressPointer{Data: &MacAddress{MacAddress: p.MacAddress}, OnConflict: macAddressOnConflict}
		pp := &PortPointer{
			Data:       &Port{System: sp, MacAddress: mp, Name: p.Name, Description: p.Description},
			OnConflict: portOnConflictMacAddressDescription,
		}
		portCache[portKey(p)] = pp
		pj := &PortJournal{Port: pp, Time: &t, Status: p.LinkStatus.String(), Speed: int(p.Speed)}
		j.Ports = append(j.Ports, pj)
	}

	for _, l := range i.LLDPs {
		var pp *PortPointer
		if p, ok := portCache[portKey(l.RemotePort)]; ok {
			pp = p
		} else {
			var sp *SystemPointer
			if s, ok := sysCache[l.RemotePort.MacAddress]; ok {
				sp = s
			} else {
				sp = &SystemPointer{Data: &System{Name: l.RemotePort.SystemName}, OnConflict: systemOnConflict}
				sysCache[l.RemotePort.MacAddress] = sp
			}
			mp := &MacAddressPointer{Data: &MacAddress{MacAddress: l.RemotePort.MacAddress}, OnConflict: macAddressOnConflict}
			pp = &PortPointer{
				Data:       &Port{System: sp, MacAddress: mp, Name: l.RemotePort.Name},
				OnConflict: portOnConflict,
			}
			portCache[portKey(l.RemotePort)] = pp
		}
		lp := &LLDPPointer{Data: &LLDP{LocalPort: portCache[portKey(l.LocalPort)], RemotePort: pp}, OnConflict: lldpOnConflict}
		lj := &LLDPJournal{LLDP: lp, Time: &t}
		j.LLDPs = append(j.LLDPs, lj)
	}

	for _, m := range i.MacAddresses {
		mp := &MacAddressPointer{Data: &MacAddress{MacAddress: m.MacAddress}, OnConflict: macAddressOnConflict}
		mj := &MacAddressJournal{MacAddress: mp, Port: portCache[portKey(m.Port)], Time: &t, Vlan: m.Vlan}
		j.MacAddresses = append(j.MacAddresses, mj)
	}

	arpCache := make(map[string]*SystemPointer)

	for _, a := range i.Arps {
		mp := &MacAddressPointer{Data: &MacAddress{MacAddress: a.MacAddress}, OnConflict: macAddressOnConflict}
		ip := &IPAddressPointer{Data: &IPAddress{IPAddress: a.IPAddress}, OnConflict: ipAddressOnConflict}
		ap := &ArpPointer{Data: &Arp{MacAddress: mp, IPAddress: ip}, OnConflict: arpOnConflict}
		aj := &ArpJournal{Arp: ap, Time: &t}
		j.Arps = append(j.Arps, aj)
		if s, ok := sysCache[a.MacAddress]; ok {
			arpCache[a.IPAddress] = s
		}
	}

	for _, r := range i.Resolves {
		ip := &IPAddressPointer{Data: &IPAddress{IPAddress: r.IPAddress}, OnConflict: ipAddressOnConflict}
		hp := &HostnamePointer{Data: &Hostname{Hostname: r.Hostname}, OnConflict: hostnameOnConflict}
		if s, ok := arpCache[r.IPAddress]; ok {
			s.Data.Hostname = hp
			s.OnConflict = systemOnConflictHostname
		}
		rp := &ResolvePointer{Data: &Resolve{IPAddress: ip, Hostname: hp}, OnConflict: resolveOnConflict}
		rj := &ResolveJournal{Resolve: rp, Time: &t}
		j.Resolves = append(j.Resolves, rj)
	}

	return j
}
