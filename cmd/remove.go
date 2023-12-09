package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
)

type removeCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newRemoveCommand(def di.ViewDef) *removeCommand {
	cmd := removeCommand{
		viewDef: def,
	}

	return &cmd
}

func (r *removeCommand) command() *cobra.Command {
	var removeCommand = &cobra.Command{
		Use:      "remove [selectors...]",
		Short:    "Removes all matching tasks permanently",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     r.remove,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	removeCommand.Flags().BoolVarP(&r.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(removeCommand, &r.qql, &r.rng, &r.str)
	return removeCommand
}

func (r *removeCommand) remove(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(r.viewDef.Query, args, r.qql, r.rng, r.str)
	if err != nil {
		return err
	}
	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !r.all {
		confirmedSelection, err = view.NewSelection(selection).Run()
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Println("no matches")
		return nil
	}

	for _, t := range confirmedSelection {
		if err := list.Remove(list.LineOf(t)); err != nil {
			return err
		}
	}
	view.NewSuccessMessage("Removed", list, confirmedSelection).Run()
	return nil
}
