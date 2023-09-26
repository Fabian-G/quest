package cmdutil

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

func AskRetry(cmd *cobra.Command, err error) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "your changes are invalid: %s\n", err)
	fmt.Fprint(cmd.OutOrStdout(), "Retry? (Y/n) ")
	var answer string
	fmt.Fscanln(cmd.InOrStdin(), &answer)
	return strings.ToLower(answer) != "n"
}

func PrintSuccessMessage(operation string, selection []*todotxt.Item) {
	switch len(selection) {
	case 0:
		fmt.Println("nothing to do")
	case 1:
		fmt.Printf("%s item: %s\n", operation, selection[0].Description())
	default:
		fmt.Printf("%s %d items\n", operation, len(selection))
	}
}
