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

var doneFunc qselect.Func = func(l *todotxt.List, i *todotxt.Item) bool {
	return i.Done()
}

type archiveCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newArchiveCommand(def di.ViewDef) *archiveCommand {
	cmd := archiveCommand{
		viewDef: def,
	}

	return &cmd
}

func (a *archiveCommand) command() *cobra.Command {
	var archiveCommand = &cobra.Command{
		Use:      "archive [selectors...]",
		Short:    "Archives all matching tasks (this will only ever match done tasks)",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList, cmdutil.LoadDoneList),
		RunE:     a.archive,
		PostRunE: cmdutil.Steps(cmdutil.SaveDoneList, cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(archiveCommand, &a.qql, &a.rng, &a.str, &a.all)
	return archiveCommand
}

func (a *archiveCommand) archive(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	doneList := cmd.Context().Value(cmdutil.DoneListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(a.viewDef.Query, args, a.qql, a.rng, a.str)
	if err != nil {
		return err
	}
	selection := qselect.And(doneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !a.all {
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
		if err := doneList.Add(t); err != nil {
			return err
		}
	}

	view.NewSuccessMessage("Archived", list, confirmedSelection).Run()
	return nil
}
