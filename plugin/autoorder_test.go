package plugin_test

import (
	"testing"

	"github.com/Fabian-G/todotxt/config"
	"github.com/Fabian-G/todotxt/plugin"
	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_AutoOrder(t *testing.T) {
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("T4")),
		todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("T3")),
		todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithDescription("T1")),
		todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("T2")),
	)
	list.AddHook(plugin.NewAutoorder(list, config.Data{}))

	list.Get(0).Complete()

	assert.Equal(t, list.Get(0).Description(), "T2")
	assert.Equal(t, list.Get(1).Description(), "T3")
	assert.Equal(t, list.Get(2).Description(), "T1")
	assert.Equal(t, list.Get(3).Description(), "T4")
}
