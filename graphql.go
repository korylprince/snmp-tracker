package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/korylprince/go-graphql-ws"
	"github.com/korylprince/snmp-tracker/snmp"
)

const gqlReadSystems = `
	query read_systems {
	  system(where: {connection_id: {_is_null: false}, hostname: {hostname: {_neq: ""}}, port: {_neq: 0}}) {
		hostname {
		  hostname
		}
		port
		connection {
		  transport
		  community
		  timeout
		  retries
		  max_oids
		  max_repetitions
		  msg_flags
		  security_model
		  auth_protocol
		  username
		  password
		  priv_protocol
		  priv_password
		}
	  }
	}
`

const gqlInsertJournal = `
mutation insert_journals(
  $ports: [port_journal_insert_input!]!,
  $lldps: [lldp_journal_insert_input!]!,
  $mac_addresses: [mac_address_journal_insert_input!]!,
  $arps: [arp_journal_insert_input!]!,
  $resolves: [resolve_journal_insert_input!]!
) {
  insert_port_journal(objects: $ports) {
    affected_rows
  }
  insert_lldp_journal(objects: $lldps) {
    affected_rows
  }
  insert_mac_address_journal(objects: $mac_addresses) {
    affected_rows
  }
  insert_arp_journal(objects: $arps) {
    affected_rows
  }
  insert_resolve_journal(objects: $resolves) {
    affected_rows
  }
}
`

type Option func(*GraphQLConn)

func WithDebugFile(path string) Option {
	return func(c *GraphQLConn) {
		c.debugPath = path
	}
}

// GraphQLConn is a GraphQL websocket connection
type GraphQLConn struct {
	conn      *graphql.Conn
	mu        *sync.Mutex
	debugPath string
}

// NewGraphQLConn returns a new GraphQLConn
func NewGraphQLConn(endpoint, adminSecret, apiSecret string, opts ...Option) (*GraphQLConn, error) {
	headers := make(http.Header)
	if adminSecret != "" {
		headers.Add("X-Hasura-Admin-Secret", adminSecret)
	} else if apiSecret != "" {
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", apiSecret))
		headers.Add("X-Authorization-Type", "API-Key")
	}

	conn, _, err := graphql.DefaultDialer.Dial(endpoint, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect: %w", err)
	}

	gc := &GraphQLConn{conn: conn, mu: new(sync.Mutex)}

	for _, opt := range opts {
		opt(gc)
	}

	conn.SetCloseHandler(func(code int, text string) {
		log.Printf("WARNING: Websocket closed unexpectedly: (%d) %s\n", code, text)
		gc.mu.Lock()
		for {
			time.Sleep(10 * time.Second)
			c, err := NewGraphQLConn(endpoint, adminSecret, apiSecret)
			if err != nil {
				log.Println("WARNING: Unable to connect to GraphQL endpoint:", err)
				continue
			}
			gc.conn = c.conn
			gc.mu.Unlock()
			return
		}
	})

	return gc, nil
}

// ReadSystems reads the Systems from the connection
func (c *GraphQLConn) ReadSystems() ([]*snmp.System, error) {
	type response struct {
		System []*struct {
			Hostname struct {
				Hostname string `json:"hostname"`
			} `json:"hostname"`
			Port             uint16                 `json:"port"`
			ConnectionConfig *snmp.ConnectionConfig `json:"connection"`
		} `json:"system"`
	}

	var q = &graphql.MessagePayloadStart{
		Query: gqlReadSystems,
	}

	c.mu.Lock()
	payload, err := c.conn.Execute(context.Background(), q)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("Unable to execute query: %w", err)
	}

	if len(payload.Errors) > 0 {
		return nil, fmt.Errorf("Unable to execute query: %w", payload.Errors)
	}

	resp := new(response)

	if err = json.Unmarshal(payload.Data, resp); err != nil {
		return nil, fmt.Errorf("Unable to decode response: %w", err)
	}

	systems := make([]*snmp.System, 0, len(resp.System))
	for _, s := range resp.System {
		systems = append(systems, &snmp.System{
			Hostname:         s.Hostname.Hostname,
			Port:             s.Port,
			ConnectionConfig: s.ConnectionConfig,
		})
	}

	return systems, nil
}

// InsertJournal submits the Journal
func (c *GraphQLConn) InsertJournal(j *Journal) (int, error) {
	type response struct {
		InsertPortJournal struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_port_journal"`
		InsertLLDPJournal struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_lldp_journal"`
		InsertMacAddressJournal struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_mac_address_journal"`
		InsertArpJournal struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_arp_journal"`
		InsertResolveJournal struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_resolve_journal"`
	}

	var q = &graphql.MessagePayloadStart{
		Query: gqlInsertJournal,
		Variables: map[string]interface{}{
			"ports":         j.Ports,
			"lldps":         j.LLDPs,
			"mac_addresses": j.MacAddresses,
			"arps":          j.Arps,
			"resolves":      j.Resolves,
		},
	}

	if c.debugPath != "" {
		f, err := os.Create(c.debugPath)
		if err != nil {
			log.Println("WARNING: Unable to create debug file:", err)
		} else {
			defer f.Close()
			if err = json.NewEncoder(f).Encode(q); err != nil {
				log.Println("WARNING: Unable to write debug file:", err)
			}
		}
	}

	c.mu.Lock()
	payload, err := c.conn.Execute(context.Background(), q)
	c.mu.Unlock()
	if err != nil {
		return 0, fmt.Errorf("Unable to execute query: %w", err)
	}

	if len(payload.Errors) > 0 {
		return 0, fmt.Errorf("Unable to execute query: %w", payload.Errors)
	}

	resp := new(response)

	if err = json.Unmarshal(payload.Data, resp); err != nil {
		return 0, fmt.Errorf("Unable to decode response: %w", err)
	}

	return resp.InsertPortJournal.Rows +
			resp.InsertLLDPJournal.Rows +
			resp.InsertMacAddressJournal.Rows +
			resp.InsertArpJournal.Rows +
			resp.InsertResolveJournal.Rows,
		nil
}
