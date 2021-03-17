package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	mutation insert_journal($journal: journal_insert_input!) {
	  insert_journal_one(object: $journal) {
		id
	  }
	}
`

type GraphQLConn struct {
	conn *graphql.Conn
}

func NewGraphQLConn(endpoint, secret string) (*GraphQLConn, error) {
	headers := make(http.Header)
	if secret != "" {
		headers.Add("X-Hasura-Admin-Secret", secret)
	}

	conn, _, err := graphql.DefaultDialer.Dial(endpoint, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect: %w", err)
	}

	return &GraphQLConn{conn: conn}, nil
}

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

	payload, err := c.conn.Execute(context.Background(), q)
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

func (c *GraphQLConn) InsertJournal(j *Journal) (int, error) {
	type response struct {
		InsertJournal struct {
			ID int `json:"id"`
		} `json:"insert_journal_one"`
	}

	var q = &graphql.MessagePayloadStart{
		Query: gqlInsertJournal,
		Variables: map[string]interface{}{
			"journal": j,
		},
	}

	payload, err := c.conn.Execute(context.Background(), q)
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

	return resp.InsertJournal.ID, nil
}
