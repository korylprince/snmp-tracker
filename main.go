package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/korylprince/ipscan/resolve"
)

func main() {
	config := new(Config)
	if err := envconfig.Process("", config); err != nil {
		log.Fatalln("ERROR: Unable to process configuration:", err)
	}

	resolver := resolve.NewService(config.Resolvers, config.ResolveBuffers)

	var opts []Option

	if config.DebugPath != "" {
		opts = append(opts, WithDebugFile(config.DebugPath))
	}

	log.Println("Connecting to", config.GraphQLEndpoint)
	conn, err := NewGraphQLConn(config.GraphQLEndpoint, config.GraphQLAdminSecret, config.GraphQLAPISecret, opts...)
	if err != nil {
		log.Fatalln("ERROR: Unable to connect to GraphQL endpoint:", err)
	}

	for {
		systems, err := conn.ReadSystems()
		if err != nil {
			log.Println("WARNING: Unable to read systems:", err)
			time.Sleep(config.PollInterval)
			continue
		}

		log.Printf("INFO: Getting information from %d systems\n", len(systems))

		info := GetInfo(resolver, systems, config.SNMPWorkers)

		j := Translate(info)

		log.Println("INFO: Inserting information into database")
		rows, err := conn.InsertJournal(j)
		if err != nil {
			log.Println("WARNING: Unable to insert information:", err)
		} else {
			log.Println("INFO:", rows, "rows inserted")
		}

		time.Sleep(config.PollInterval)
	}
}
