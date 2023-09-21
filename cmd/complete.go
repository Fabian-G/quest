package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
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
	all     bool
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
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     c.complete,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	completeCmd.Flags().BoolVarP(&c.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(completeCmd, &c.qql, &c.rng, &c.str)
	return completeCmd
}

func (c *completeCommand) complete(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(c.viewDef.DefaultQuery, args, c.qql, c.rng, c.str)
	if err != nil {
		return err
	}
	selection := qselect.And(notDoneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !c.all {
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
		if err := t.Complete(); err != nil {
			return err
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Successfully completed %d items\n", len(confirmedSelection))
	return nil
}
