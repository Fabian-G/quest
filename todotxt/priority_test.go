package todotxt_test

import (
	"testing"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func TestPriority_ToString(t *testing.T) {
	tests := []struct {
		name string
		p    todotxt.Priority
		want string
	}{
		{"None returns an empty string", todotxt.PrioNone, ""},
		{"A", todotxt.PrioA, "(A)"},
		{"Z", todotxt.PrioZ, "(Z)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestPriority_FromString(t *testing.T) {
	tests := []struct {
		name string
		want todotxt.Priority
		p    string
	}{
		{"A", todotxt.PrioA, "(A)"},
		{"Z", todotxt.PrioZ, "(Z)"},
		{"None", todotxt.PrioNone, "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prio, err := todotxt.PriorityFromString(tt.p)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, prio)
		})
	}
}
