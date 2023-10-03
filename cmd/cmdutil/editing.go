package cmdutil

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

func AskRetry(cmd *cobra.Command, err error) bool {
	fmt.Printf("your changes are invalid: %s\n", err)
	fmt.Print("Retry? (Y/n) ")
	var answer string
	_, _ = fmt.Fscanln(cmd.InOrStdin(), &answer)
	return strings.ToLower(answer) != "n"
}

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
