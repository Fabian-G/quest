package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

var doneFunc qselect.Func = func(l *todotxt.List, i *todotxt.Item) bool {
	return i.Done()
}

type archiveCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newArchiveCommand(def config.ViewDef) *archiveCommand {
	cmd := archiveCommand{
		viewDef: def,
	}

	return &cmd
}

func (a *archiveCommand) command() *cobra.Command {
	var archiveCommand = &cobra.Command{
		Use:      "archive",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList, cmdutil.LoadDoneList),
		RunE:     a.archive,
		PostRunE: cmdutil.Steps(cmdutil.SaveDoneList, cmdutil.SaveList),
	}
	archiveCommand.Flags().BoolVarP(&a.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(archiveCommand, &a.qql, &a.rng, &a.str)
	return archiveCommand
}

func (a *archiveCommand) archive(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	doneList := cmd.Context().Value(cmdutil.DoneListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(a.viewDef.DefaultQuery, args, a.qql, a.rng, a.str)
	if err != nil {
		return err
	}
	selection := qselect.And(doneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !a.all {
		confirmedSelection, err = cmdutil.ConfirmSelection(selection)
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}

	for _, t := range confirmedSelection {
		if err := list.Remove(list.IndexOf(t)); err != nil {
			return err
		}
		if err := doneList.Add(t); err != nil {
			return err
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Successfully archived %d items\n", len(confirmedSelection))
	return nil
}
