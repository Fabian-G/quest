package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
)

var notDoneFunc qselect.Func = func(l *todotxt.List, i *todotxt.Item) bool {
	return !i.Done()
}

type completeCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newCompleteCommand(def di.ViewDef) *completeCommand {
	cmd := completeCommand{
		viewDef: def,
	}

	return &cmd
}

func (c *completeCommand) command() *cobra.Command {
	var completeCmd = &cobra.Command{
		Use:      "complete [selectors...]",
		Short:    "Completes all matching tasks",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     c.complete,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(completeCmd, &c.qql, &c.rng, &c.str, &c.all)
	return completeCmd
}

func (c *completeCommand) complete(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(c.viewDef.Query, args, c.qql, c.rng, c.str)
	if err != nil {
		return err
	}
	selection := qselect.And(notDoneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !c.all {
		confirmedSelection, err = view.NewSelection(selection).Run()
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Print("no matches")
		return nil
	}

	for _, t := range confirmedSelection {
		if err := t.Complete(); err != nil {
			return err
		}
	}
	view.NewSuccessMessage("Completed", list, confirmedSelection).Run()
	return nil
}
