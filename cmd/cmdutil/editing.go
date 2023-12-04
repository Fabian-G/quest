package cmdutil

import (
	"fmt"

	"github.com/Fabian-G/quest/todotxt"
)

func PrintSuccessMessage(operation string, list *todotxt.List, selection []*todotxt.Item) {
	switch len(selection) {
	case 0:
		fmt.Println("nothing to do")
	case 1:
		fmt.Printf("%s item #%d: %s\n", operation, list.LineOf(selection[0]), selection[0].Description())
	default:
		fmt.Printf("%s %d items\n", operation, len(selection))
	}
}
