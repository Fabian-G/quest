package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type setCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newSetCommand(def config.ViewDef) *setCommand {
	cmd := setCommand{
		viewDef: def,
	}

	return &cmd
}

func (s *setCommand) command() *cobra.Command {
	var setCommand = &cobra.Command{
		Use:      "set",
		Args:     cobra.MinimumNArgs(2),
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     s.set,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	setCommand.Flags().BoolVarP(&s.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(setCommand, &s.qql, &s.rng, &s.str)
	return setCommand
}

func (s *setCommand) set(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	tag := args[0]
	value := args[1]

	selector, err := cmdutil.ParseTaskSelection(s.viewDef.DefaultQuery, args[2:], s.qql, s.rng, s.str)
	if err != nil {
		return err
	}

	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !s.all {
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
		if err := t.SetTag(tag, value); err != nil {
			return err
		}
	}
	if value == "" {
		cmdutil.PrintSuccessMessage(fmt.Sprintf("Removed tag \"%s\" from", tag), list, confirmedSelection)
	} else {
		cmdutil.PrintSuccessMessage(fmt.Sprintf("Set tag \"%s\" to \"%s\" on", tag, value), list, confirmedSelection)
	}
	return nil
}
