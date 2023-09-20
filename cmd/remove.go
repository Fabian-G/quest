package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type removeCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
}

func newRemoveCommand(def config.ViewDef) *removeCommand {
	cmd := removeCommand{
		viewDef: def,
	}

	return &cmd
}

func (r *removeCommand) command() *cobra.Command {
	var removeCommand = &cobra.Command{
		Use:      "remove",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		PreRunE:  steps(loadList),
		RunE:     r.remove,
		PostRunE: steps(saveList),
	}
	registerSelectionFlags(removeCommand, &r.qql, &r.rng, &r.str)
	return removeCommand
}

func (r *removeCommand) remove(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(listKey).(*todotxt.List)
	selector, err := parseTaskSelection(r.viewDef.DefaultQuery, args, r.qql, r.rng, r.str)
	if err != nil {
		return err
	}
	selection := selector.Filter(list)
	confirmedSelection, err := confirmSelection(selection)
	if err != nil {
		return err
	}

	for _, t := range confirmedSelection {
		if err := list.Remove(list.IndexOf(t)); err != nil {
			return err
		}
	}
	fmt.Printf("Successfully removed %d items\n", len(confirmedSelection))
	return nil
}
