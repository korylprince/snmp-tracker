package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/korylprince/go-graphql-ws"
)

type system struct {
	Name         string `json:"name"`
	Hostname     string `json:"hostname"`
	ConnectionID int    `json:"connection_id"`
	Port         *int   `json:"port,omitempty"`
}

func parseSystems(path string) ([]*system, error) {
	type parse struct {
		Systems []*system `json:"systems"`
	}

	p := new(parse)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file: %w", err)
	}
	dec := json.NewDecoder(f)
	if err = dec.Decode(p); err != nil {
		return nil, fmt.Errorf("Unable to parse file: %w", err)
	}

	return p.Systems, nil
}

type Upsert struct {
	Constraint    string   `json:"constraint"`
	UpdateColumns []string `json:"update_columns"`
}

var hostnameOnConflict = &Upsert{Constraint: "unique_hostname", UpdateColumns: []string{"hostname"}}

type Hostname struct {
	Hostname string `json:"hostname"`
}

type HostnamePointer struct {
	Data       *Hostname `json:"data"`
	OnConflict *Upsert   `json:"on_conflict"`
}

type System struct {
	Name         string           `json:"name"`
	Hostname     *HostnamePointer `json:"hostname"`
	ConnectionID int              `json:"connection_id"`
	Port         *int             `json:"port,omitempty"`
}

func translate(sys []*system) []*System {
	systems := make([]*System, 0, len(sys))
	for _, s := range sys {
		systems = append(systems, &System{
			Name: s.Name,
			Hostname: &HostnamePointer{
				Data:       &Hostname{Hostname: s.Hostname},
				OnConflict: hostnameOnConflict,
			},
			ConnectionID: s.ConnectionID,
			Port:         s.Port,
		})
	}

	return systems
}

const gqlInsert = `
	mutation insert_systems($systems: [system_insert_input!]!) {
	  insert_system(objects: $systems, on_conflict: {constraint: unique_system_name, update_columns: [name, hostname_id, connection_id, port]}) {
		affected_rows
	  }
	}
`

func insertSystems(endpoint, secret string, systems []*System) (int, error) {
	type response struct {
		InsertSystem struct {
			Rows int `json:"affected_rows"`
		} `json:"insert_system"`
	}

	headers := make(http.Header)
	headers.Add("X-Hasura-Admin-Secret", secret)

	conn, _, err := graphql.DefaultDialer.Dial(endpoint, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("Unable to connect: %w", err)
	}

	var q = &graphql.MessagePayloadStart{
		Query: gqlInsert,
		Variables: map[string]interface{}{
			"systems": systems,
		},
	}

	payload, err := conn.Execute(context.Background(), q)
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

	return resp.InsertSystem.Rows, nil
}

func main() {
	path := flag.String("path", "", "path to systems.json")
	url := flag.String("url", "", "connection url, e.g. wss://example.com/v1/graphql")
	secret := flag.String("secret", "", "Hasura admin secret")
	flag.Parse()

	if *path == "" {
		fmt.Println("-path must be set")
		os.Exit(-1)
	}
	if *url == "" {
		fmt.Println("-url must be set")
		os.Exit(-1)
	}
	if *secret == "" {
		fmt.Println("-secret must be set")
		os.Exit(-1)
	}

	sys, err := parseSystems(*path)
	if err != nil {
		fmt.Println("Unable to parse systems:", err)
		os.Exit(-1)
	}

	systems := translate(sys)

	n, err := insertSystems(*url, *secret, systems)
	if err != nil {
		fmt.Println("Unable to insert systems:", err)
		os.Exit(-1)
	}

	fmt.Println("Affected", n, "rows")
}
