package main

import "time"

// Config configures snmp-tracker
type Config struct {
	GraphQLEndpoint    string `required:"true"`
	GraphQLAdminSecret string
	GraphQLAPISecret   string
	SNMPWorkers        int           `default:"10"`
	Resolvers          int           `default:"16"`
	ResolveBuffers     int           `default:"1024"`
	PollInterval       time.Duration `default:"30m"`
}
