package datalog

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReset(t *testing.T) {
	t.Parallel()

	dat := newData()
	dat.Hits = 10
	dat.Bytes = 100
	dat.Resources = map[string]*data{}

	dat.reset()

	if !reflect.DeepEqual(dat, &data{}) {
		t.Fatalf("expected empty data, got: %v", dat)
	}
}

func TestProcess(t *testing.T) {
	t.Parallel()

	type input struct {
		remoteHost string
		authUser   string
		method     string
		resource   string
		status     uint16
		bytes      uint64
	}
	mkInput := func(authUser, method, resource string, status uint16, bytes uint64) input {
		return input{
			authUser: authUser,
			method:   method,
			resource: resource,
			status:   status,
			bytes:    bytes,
		}
	}

	tests := []struct {
		in  []input
		exp *data
	}{
		{
			in: []input{{authUser: "yo", bytes: 10}},
			exp: &data{
				Hits:     1,
				Bytes:    10,
				Methods:  map[string]int64{},
				Users:    map[string]int64{"yo": 1},
				Statuses: map[uint16]int64{},
			},
		},

		{
			in: []input{mkInput("user", "method", "resource", 200, 10)},
			exp: &data{
				Hits:  1,
				Bytes: 10,
				Resources: map[string]*data{
					"resource": {
						Hits:     1,
						Bytes:    10,
						Methods:  map[string]int64{"method": 1},
						Users:    map[string]int64{"user": 1},
						Statuses: map[uint16]int64{200: 1},
					},
				},
				Methods:  map[string]int64{"method": 1},
				Users:    map[string]int64{"user": 1},
				Statuses: map[uint16]int64{200: 1},
			},
		},

		{
			in: []input{
				mkInput("user1", "method1", "resource", 200, 10),
				mkInput("user2", "method2", "resource", 400, 1),
			},
			exp: &data{
				Hits:  2,
				Bytes: 11,
				Resources: map[string]*data{
					"resource": {
						Hits:  2,
						Bytes: 11,
						Methods: map[string]int64{
							"method1": 1,
							"method2": 1,
						},
						Users: map[string]int64{
							"user1": 1,
							"user2": 1,
						},
						Statuses: map[uint16]int64{
							200: 1,
							400: 1,
						},
					},
				},
				Methods: map[string]int64{
					"method1": 1,
					"method2": 1,
				},
				Users: map[string]int64{
					"user1": 1,
					"user2": 1,
				},
				Statuses: map[uint16]int64{
					200: 1,
					400: 1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			dat := newData()

			for _, in := range test.in {
				dat.process(in.authUser, in.method, in.resource, in.status, in.bytes)
			}

			assert.Equal(t, test.exp, dat)
		})
	}
}
