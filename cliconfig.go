package main

import "time"

// CliConfig defines available cli configurations
type CliConfig struct {
	ghRepo          string `required:"true"`
	project         string `required:"true"`
	category        string
	sleepTime       time.Duration
	debug           bool
	ticketTemplate  string
	commentTemplate string
}
