package datalog

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlert(t *testing.T) {
	t.Parallel()

	interval := 5 * time.Millisecond

	tests := []struct {
		threshold uint64
		many      int
		post      func(*D)
	}{
		{threshold: 5, many: 10},
		{threshold: 5, many: 2},
		{threshold: 5, many: 2, post: func(d *D) {
			time.Sleep(2*interval + 1*time.Millisecond)

			assert.False(t, d.wasHigh, "expected alert to have gone away")
		}},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			// open temporary log file

			f, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer func() {
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()
			err = f.Sync()
			assert.NoError(t, err)

			// configure the service

			config := DefaultConfig
			config.Fname = f.Name()
			config.TrafficInterval = interval
			config.TrafficThreshold = test.threshold

			d := NewD(WithConfig(config))
			mtx := sync.Mutex{} // to appease race detector
			go func() {
				mtx.Lock()
				defer mtx.Unlock()

				// start the service

				d.Start()
			}()

			// write the log lines

			for i := 0; i < test.many; i++ {
				_, err := f.WriteString(`127.0.0.1 - frank [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 1234`)
				assert.NoError(t, err)
				_, _ = f.WriteString("\n")
				err = f.Sync()
				assert.NoError(t, err)
			}
			d.stop()
			time.Sleep(50 * time.Millisecond)

			// check that alert was correct

			mtx.Lock()
			// alert when we generated more traffic than threshold
			assert.Equal(t, uint64(test.many) > test.threshold, d.wasHigh)
			mtx.Unlock()

			// run function that tests if alert went away

			if test.post != nil {
				test.post(d)
			}
		})
	}
}
