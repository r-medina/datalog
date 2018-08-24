package datalog

import (
	"log"
	"time"
)

// Config contains the configuration for D.
type Config struct {
	Fname string

	ReportInterval time.Duration

	TrafficInterval  time.Duration
	TrafficThreshold uint64

	Debug bool

	L *log.Logger
}

// DefaultConfig contains sane defaults for D. This is instantiated in this way
// so that it is copied and callers cannot tamper with it accidentally.
func DefaultConfig() Config {
	return Config{
		Fname:            "/var/log/access.log",
		ReportInterval:   10 * time.Second,
		TrafficInterval:  2 * time.Minute,
		TrafficThreshold: 10,
	}
}
