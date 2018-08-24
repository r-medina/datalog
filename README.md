[![build status](https://travis-ci.org/r-medina/datalog.svg?branch=master)](https://travis-ci.org/r-medina/datalog)
[![GoDoc](https://godoc.org/github.com/r-medina/datalog?status.svg)](https://godoc.org/github.com/r-medina/datalog)

# Datalog

This repository contains my golang implementation of the datadog coding challenge.

The way the code is structured is fairly simple: this directory contains the top-level package. This package contains an extremely minimal API (configuraton/starting) and the package in `cmd/datalog` contains the code that compiles to an executable.

Also in this repository, however, are various scripts for building and CI.

`datalog` comes with travis-ci integration that automatically runs `go vet`, `golint`, `go build`, and `go test` (with `race` and `cover`) on all the packages.

The build scripts help build binaries and docker images.

## Design

My intention with this code was to keep it as simple as possible.

After reading the prompt a few times, I identified that it was more important to show I can make a useful, maintainable tool than show I can build build every little component. With this in mind, rather than building a Common Log Format parser myself (I kind of love building parsers) I imported one I found on the internet: `"github.com/tmcb/clfmon/clf"`. Another component I imported was the file tailing. For this, I used `"github.com/hpcloud/tail"`. Finally, for making cli apps in go, I used `"gopkg.in/alecthomas/kingpin.v2"` (which is an amazing peice of software).

The one exported type `D` only supports instantiation and a function to start it up. The constructor, `NewD()`, allows the caller to specify

1. The name of the log file (defaults to `/var/log/access.log`)
2. The reporting interval (ie how often the report displays)
3. The traffic alert time window (ie the window of time within which traffic has to exceed an amount to generate an alert)
4. Whether or not debug logs should print
5. The `*log.Logger` `D` uses.

There is also a default configuration available.

Almost everything else is kept from package users.

The component of `D` is the method to start it. `Start()` tails the file specified (it has to exist prior to running although this can be easily changed) and generates the reports or alerts. Because reorting is so infrequent (potentially) relative to traffic and because file writes can only be sequential (to make sense), I kept everything single threaded. If we find that traffic is blocked too long while the report is printed (unlikely), we could spin off the printing into a goroutine and use synchronization primitives (perhaps a read/write mutex) to protect the data (or make a quick copy and keep it moving).

The data that goes into the printed report is hidden behind an unexported type called `data` which, as of the time of this writing, is defined as:

```go
// data represents the data processed for reporting.
// This type is mildly recursive, as the `sub` field contains information about
// a specific resource.
type data struct {
	Hits      int64            `json:"hits"`
	Resources map[string]*data `json:"resources,omitempty"`
	Methods   map[string]int64 `json:"methods,omitempty"`
	Users     map[string]int64 `json:"users,omitempty"`
	Statuses  map[uint16]int64 `json:"statuses,omitempty"`
	Bytes     uint64           `json:"bytes"`

	sync.RWMutex
}
```

*I made the explicit design decision to omit some of the fields that CLF contains.* They can be added with relative ease, but for the sake of making something quick and useful (based on the logs in the prompt) I only focus on the fields mentioned above.

While this type is recursive (the `Resources` field is a map of strings to another `*data`), it has at most 2 levels. The highest level contains an overview of all requests that came in. It keeps a running total of all the hits in `Hits` and a total of bytes sent in `Bytes`. The other fields (other than `Resources` which I will cover later) are histograms which map a value to the amount of times it was seen. That is to say, `Users` will have a mapping of user names to the number of times they called the service (see below).

`Resources` is a bit more interesting. Rather than using the section definition in the prompt, I merely index on the resource returned from the CLF parser. This decision made more sense with the logs in the prompt, as they contain resources like `/report` and `/api/user` - it did not make sense to me to further parse these. Each resource will have another `*data` object associated with it although it will not point to any children. This is to say each resource will have information about users, statuses, bytes, hits, and methods.

The program prints reorts to stdout in the form of human-readable JSON.

The report output looks like:

```
{
    "hits": 35,
    "resources": {
        "/api/user": {
            "hits": 26,
            "methods": {
                "GET": 26
            },
            "users": {
                "frank": 9,
                "jill": 9,
                "mary": 8
            },
            "statuses": {
                "200": 26
            },
            "bytes": 32084
        },
        "/report": {
            "hits": 9,
            "methods": {
                "GET": 9
            },
            "users": {
                "james": 9
            },
            "statuses": {
                "200": 9
            },
            "bytes": 11106
        }
    },
    "methods": {
        "GET": 35
    },
    "users": {
        "frank": 9,
        "james": 9,
        "jill": 9,
        "mary": 8
    },
    "statuses": {
        "200": 35
    },
    "bytes": 43190
}
```

All other output (alerts and logs) go to stderr. This allows for the output to be piped into a program that perhaps makes visualizations baded on the JSON output.

My alerts work a little differently than mandated by the prompt. As soon as the traffic excedes some threshold value (within the specified time window), an alert is generated and printed to stderr. The alert is deescalated when another time window goes by wherein the traffic is less than the threshold value. There can be some lag here though as the window that generated the alert needs to finish and then another one needs to ellapse before the alert is deescalated. This, however, is not terrible, as there will be more certainty that the alert condition is no longer met if there is a longer period of time to sample.

## Running

The compiled binary for the `datalog` service has the following usage:

```
usage: datalog [<flags>]

see info about logs

Flags:
  --help                   Show context-sensitive help (also try --help-long and --help-man).
  --file-name="/var/log/access.log"  
                           name of log file to stream
  --report-interval=10s    amount of time before stats are reported
  --traffic-interval=2m0s  time over whch to measure high traffic
  --traffic-threshold=10   amount of traffic an endpoint receives before reporting high traffic (over interval)
  --debug                  if debug is enabled
  --pprof-addr=PPROF-ADDR  address for running pprof tools
```

These are all configurable through environmental variables (so that a service definition just has to run the executable) as well as command-line flags.

Although one can build the executables and run it, it is preferable to use docker. In order to do this:

```sh
docker pull rxaxm/datalog
docker run datalog --pprof-addr=:9999
```

The docker container expects the log to be in `/var/log/access.log`, so, in order to use this, one would have to execute a command that appends to the file in the container (the container could also mount `/var/log/` with some additional configuration).

## Building

Although I did no fancy go vendoring, if you want to build this locally, a `go get -u ./... && go build ./...` should do the trick. If you want the run-able binary, you need to go into `cmd/datalog` and run `go build`.

The Makefile in this repository can help you build the docker image. Running `make deps && make` should build the binaries (executable and testing binary) for linux and make an image with those binaries in `/go`.

## Testing

`go run ./...` will run all the tests assuming `go get -u ./...` works.

There is a better script, however (`scrips/test.sh`), for running tests and seeing information about coverage and race conditions.

Furthermore, the `.travis.yml` file included with this repo should cause tests to run automatically in travis.

If you want to test using the docker image, make sure to run the docker container (eg `docker run datalog`), then, after finding the container ID with `docker ps`, `docker exec -it $CONTAINER_ID /go/datalog.linux-amd64.test -test.v` (the compiled go testing binary is in the container).
