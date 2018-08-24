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

// DefaultConfig contains sane defaults for D.
var DefaultConfig = Config{
	Fname:            "/var/log/access.log",
	ReportInterval:   10 * time.Second,
	TrafficInterval:  2 * time.Minute,
	TrafficThreshold: 10,
}
