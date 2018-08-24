package datalog

import (
	"encoding/json"
	"os"
	"sync"
)

var encoder = json.NewEncoder(os.Stdout)

func init() {
	encoder.SetIndent("", "    ")
}

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

func newData() *data {
	return &data{}
}

func (dat *data) reset() {
	dat.Hits = 0
	dat.Resources = nil
	dat.Methods = nil
	dat.Users = nil
	dat.Statuses = nil
	dat.Bytes = 0
}

func (dat *data) process(authUser, method, resource string, status uint16, bytes uint64) {
	// Because `data` is a somewhat recursive type, I include this helper
	// such that we can do the operations on the node without recursing
	// into the children.
	helper := func(dat *data) {
		// these are left blank so that the output omits them
		if dat.Methods == nil {
			dat.Methods = map[string]int64{}
			dat.Users = map[string]int64{}
			dat.Statuses = map[uint16]int64{}
		}

		dat.Hits++
		dat.Bytes += bytes
		if method != "" {
			dat.Methods[method]++
		}
		if authUser != "" {
			dat.Users[authUser]++
		}
		if status != 0 {
			dat.Statuses[status]++
		}
	}
	helper(dat)

	if resource == "" {
		return
	}

	if dat.Resources == nil {
		dat.Resources = map[string]*data{}
	}
	sub, ok := dat.Resources[resource]
	if ok {
		helper(sub)
		return
	}

	sub = newData()
	helper(sub)

	dat.Resources[resource] = sub
}

func (dat *data) print() {
	_ = encoder.Encode(dat)
}
