package snmp

import (
	"time"

	"github.com/gosnmp/gosnmp"
)

// ConnectionConfig is configuration for a connection
type ConnectionConfig struct {
	Transport      string                     `json:"transport"`
	Community      string                     `json:"community"`
	Timeout        time.Duration              `json:"timeout"`
	Retries        int                        `json:"retries"`
	MaxOIDs        int                        `json:"max_oids"`
	MaxRepetitions uint32                     `json:"max_repetitions"`
	MsgFlags       gosnmp.SnmpV3MsgFlags      `json:"msg_flags"`
	SecurityModel  gosnmp.SnmpV3SecurityModel `json:"security_model"`
	AuthProtocol   gosnmp.SnmpV3AuthProtocol  `json:"auth_protocol"`
	Username       string                     `json:"username"`
	AuthPassword   string                     `json:"password"`
	PrivProtocol   gosnmp.SnmpV3PrivProtocol  `json:"priv_protocol"`
	PrivPassword   string                     `json:"priv_password"`
}

// New returns a new SNMP configuration
func (c *ConnectionConfig) New(host string, port uint16) *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Target:         host,
		Port:           port,
		Version:        gosnmp.Version3,
		Transport:      c.Transport,
		Community:      c.Community,
		Timeout:        time.Second * c.Timeout,
		Retries:        c.Retries,
		MaxOids:        c.MaxOIDs,
		MaxRepetitions: c.MaxRepetitions,
		MsgFlags:       c.MsgFlags,
		SecurityModel:  c.SecurityModel,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			AuthenticationProtocol:   c.AuthProtocol,
			UserName:                 c.Username,
			AuthenticationPassphrase: c.AuthPassword,
			PrivacyProtocol:          c.PrivProtocol,
			PrivacyPassphrase:        c.PrivPassword,
		},
	}
}
