package cmdutil

import (
	"fmt"

	"github.com/Fabian-G/quest/qselect"
	"github.com/spf13/cobra"
)

func RegisterSelectionFlags(cmd *cobra.Command, qql *[]string, rng *[]string, str *[]string, all *bool) {
	cmd.Flags().StringArrayVarP(qql, "qql", "q", nil, "QQL Query")
	cmd.Flags().StringArrayVarP(rng, "range", "r", nil, "Range Query")
	cmd.Flags().StringArrayVarP(str, "word", "w", nil, "Case-Insensitive String Search")
	if all != nil {
		cmd.Flags().BoolVarP(all, "all", "a", false, "Don't ask for confirmation when multiple results match")
	}
}

func ParseTaskSelection(defaultQuery string, guess, qqlSearch, rngSearch, stringSearch []string) (qselect.Func, error) {
	selectors := make([]qselect.Func, 0)
	q, err := qselect.CompileQQL(defaultQuery)
	if err != nil {
		return nil, fmt.Errorf("config file contains invalid query: %w", err)
	}
	selectors = append(selectors, q)
	for _, arg := range guess {
		q, err := qselect.CompileQuery(arg)
		if err != nil {
			return nil, fmt.Errorf("could not compile query %s. Try using -q,-r or -s explicitly instead of positional args: %w", arg, err)
		}
		selectors = append(selectors, q)
	}
	for _, f := range qqlSearch {
		q, err := qselect.CompileQQL(f)
		if err != nil {
			return nil, fmt.Errorf("could not compile QQL query %s: %w", f, err)
		}
		selectors = append(selectors, q)
	}
	for _, r := range rngSearch {
		q, err := qselect.CompileRange(r)
		if err != nil {
			return nil, fmt.Errorf("could not compile range query %s: %w", r, err)
		}
		selectors = append(selectors, q)
	}
	for _, s := range stringSearch {
		q, err := qselect.CompileWordSearch(s)
		if err != nil {
			return nil, fmt.Errorf("could not compile string search query %s: %w", s, err)
		}
		selectors = append(selectors, q)
	}

	return qselect.And(selectors...), nil
}
