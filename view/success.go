package view

import (
	"fmt"

	"github.com/Fabian-G/quest/todotxt"
)

type SuccessMessage struct {
	operation string
	list      *todotxt.List
	selection []*todotxt.Item
}

func NewSuccessMessage(operation string, list *todotxt.List, selection []*todotxt.Item) *SuccessMessage {
	msg := SuccessMessage{
		operation: operation,
		list:      list,
		selection: selection,
	}
	return &msg
}

func (s SuccessMessage) Run() {
	switch len(s.selection) {
	case 0:
		fmt.Println("nothing to do")
	case 1:
		fmt.Printf("%s item #%d: %s\n", s.operation, s.list.LineOf(s.selection[0]), s.selection[0].Description())
	default:
		fmt.Printf("%s %d items\n", s.operation, len(s.selection))
	}
}
