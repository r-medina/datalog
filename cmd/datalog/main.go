package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/r-medina/datalog"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var config = struct {
	datalog.Config

	pprofAddr *net.TCPAddr
}{
	Config: datalog.DefaultConfig(),
}

var (
	app = kingpin.New("datalog", "see info about logs").
		PreAction(startPprof).Action(run).DefaultEnvars()

	l = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
)

func init() {
	app.Flag("file-name", "name of log file to stream").Default(config.Fname).StringVar(&config.Fname)
	app.Flag("report-interval", "amount of time before stats are reported").
		Default(str(config.ReportInterval)).DurationVar(&config.ReportInterval)
	app.Flag("traffic-interval", "time over whch to measure high traffic").
		Default(str(config.TrafficInterval)).DurationVar(&config.TrafficInterval)
	app.Flag("traffic-threshold", "amount of traffic an endpoint receives before reporting high traffic (over interval)").
		Default(str(config.TrafficThreshold)).Uint64Var(&config.TrafficThreshold)
	app.Flag("debug", "if debug is enabled").BoolVar(&config.Debug)

	app.Flag("pprof-addr", "address for running pprof tools").TCPVar(&config.pprofAddr)
}

func main() {
	if _, err := app.Parse(os.Args[1:]); err != nil {
		app.FatalUsage("command line parsing failed: %v", err)
	}
}

func run(_ *kingpin.ParseContext) error {
	config.L = l
	d := datalog.NewD(datalog.WithConfig(config.Config))

	l.Printf("starting datalog")
	return d.Start()
}

func startPprof(_ *kingpin.ParseContext) error {
	if config.pprofAddr == nil {
		return nil
	}

	l.Printf("running pprof server on %s", config.pprofAddr)
	go func() {
		err := http.ListenAndServe(config.pprofAddr.String(), nil)
		fatalIfError(err, "pprof server failed")
	}()

	return nil
}

func fatalIfError(err error, format string, args ...interface{}) {
	if err != nil {
		if format != "" {
			format += ": "
		}
		l.Fatalf(format+"%v", append(args, err)...)
	}
}

func str(v interface{}) string { return fmt.Sprintf("%v", v) }
