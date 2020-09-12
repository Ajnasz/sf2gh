package main

import "time"

// CliConfig defines available cli configurations
type CliConfig struct {
	ghRepo          string `required:"true"`
	project         string `required:"true"`
	dbFile          string
	category        string
	sleepTime       time.Duration
	debug           bool
	ticketTemplate  string
	commentTemplate string
}
