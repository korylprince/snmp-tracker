package main

import "time"

type Config struct {
	GraphQLEndpoint  string `required:"true"`
	GraphQLAPISecret string
	SNMPWorkers      int           `default:"10"`
	Resolvers        int           `default:"16"`
	ResolveBuffers   int           `default:"1024"`
	PollInterval     time.Duration `default:"30m"`
}