package datalog

import (
	"fmt"
	"os"
	"time"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"
	"github.com/tmcb/clfmon/clf"
)

// D is the datalog daemon.
type D struct {
	config Config
	hits   uint64 // hits per Config.TrafficInterval
	data   *data

	wasHigh bool // keeps track of if we hit the high traffic threshold

	stopCh chan struct{}
}

// NewD instantiates a new datalog daemon.
// All values are initialized to `DefaultConfig`.
func NewD(opts ...Option) *D {
	d := &D{
		config: DefaultConfig,
		data:   newData(),
		stopCh: make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Option customizes D.
type Option func(*D)

// WithConfig allows custom config.
func WithConfig(config Config) Option {
	return func(d *D) {
		d.config = config
	}
}

// Start starts the log streaming.
func (d *D) Start() error {
	t, err := tail.TailFile(d.config.Fname, tail.Config{
		MustExist: true,
		Follow:    true,
		ReOpen:    true,
	})
	if err != nil {
		return errors.Wrap(err, "TailFile failed")
	}

	var (
		reportTick  = time.Tick(d.config.ReportInterval)
		trafficTick = time.Tick(d.config.TrafficInterval)
	)

	for {
		select {
		case <-d.stopCh:
			return nil
		case line, ok := <-t.Lines:
			if !ok /* || line.Text == "" */ {
				continue
			}

			d.hits++
			if d.hits > d.config.TrafficThreshold && !d.wasHigh {
				d.alert(
					"ALERT! Traffic in the last %v has exceeded %v hits",
					d.config.TrafficInterval, d.config.TrafficThreshold,
				)
				d.wasHigh = true
			}

			if err := line.Err; err != nil {
				d.Log("error reading line: %v", err)
				continue
			}

			if err := d.processLine(line.Text); err != nil {
				d.Log("error processing line %q: %v", line.Text, err)
			}
		case <-reportTick:
			d.printReport()
			d.data.reset()
		case <-trafficTick:
			d.Log("traffic")

			if d.wasHigh && d.hits < d.config.TrafficThreshold {
				d.alert("alert over")
				d.wasHigh = false
			}

			d.hits = 0
		}
	}
}

func (d *D) stop() {
	d.stopCh <- struct{}{}
}

func (d *D) processLine(line string) error {
	d.Log("processing %q", line)

	e, err := clf.Parse(line)
	if err != nil {
		return err
	}

	d.data.process(e.AuthUser, e.Method, e.Resource, e.Status, e.Bytes)

	return nil
}

func (d *D) printReport() {
	d.Log("reporting")

	d.data.print()
}

// Log does basic logging.
func (d *D) Log(format string, a ...interface{}) {
	if !d.config.Debug || d.config.L == nil {
		return
	}

	d.config.L.Printf(format, a...)
}

func (d *D) alert(format string, a ...interface{}) {
	if d.config.L == nil {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
		return
	}

	d.config.L.Printf(format, a...)
}
