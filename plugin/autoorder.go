package plugin

import (
	"slices"

	"github.com/Fabian-G/todotxt/config"
	"github.com/Fabian-G/todotxt/query"
	"github.com/Fabian-G/todotxt/todotxt"
)

type Autoorder struct {
	list   *todotxt.List
	config config.Data
}

func NewAutoorder(list *todotxt.List, config config.Data) *Autoorder {
	return &Autoorder{
		list:   list,
		config: config,
	}
}

func (a Autoorder) sortString() string {
	// TODO read from config
	return "+done,+creation,+description"
}

func (a Autoorder) Handle(event todotxt.ModEvent) error {
	sortFunc, err := query.SortFunc(a.sortString())
	if err != nil {
		return err
	}
	slices.SortStableFunc(a.list.Tasks(), sortFunc)
	return nil
}
