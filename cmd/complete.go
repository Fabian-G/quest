package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

var notDoneFunc qselect.Func = func(l *todotxt.List, i *todotxt.Item) bool {
	return !i.Done()
}

type completeCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
}

func newCompleteCommand(def config.ViewDef) *completeCommand {
	cmd := completeCommand{
		viewDef: def,
	}

	return &cmd
}

func (c *completeCommand) command() *cobra.Command {
	var completeCmd = &cobra.Command{
		Use:      "complete",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		PreRunE:  steps(loadList),
		RunE:     c.complete,
		PostRunE: steps(saveList),
	}
	registerSelectionFlags(completeCmd, &c.qql, &c.rng, &c.str)
	return completeCmd
}

func (c *completeCommand) complete(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(listKey).(*todotxt.List)
	selector, err := parseTaskSelection(c.viewDef.DefaultQuery, args, c.qql, c.rng, c.str)
	if err != nil {
		return err
	}
	selection := qselect.And(notDoneFunc, selector).Filter(list)
	confirmedSelection, err := confirmSelection(selection)
	if err != nil {
		return err
	}

	for _, t := range confirmedSelection {
		if err := t.Complete(); err != nil {
			return err
		}
	}
	fmt.Printf("Successfully completed %d items\n", len(confirmedSelection))
	return nil
}
