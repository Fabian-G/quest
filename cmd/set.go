package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

var setRegex = regexp.MustCompile("^[^[:space:]]+=[^[:space:]]*$")

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
		Args:     cobra.MinimumNArgs(1),
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
	selectors := make([]string, 0)
	tagOps := make([]string, 0)
	for _, arg := range args {
		if setRegex.MatchString(arg) {
			tagOps = append(tagOps, arg)
		} else {
			selectors = append(selectors, arg)
		}
	}
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)

	selector, err := cmdutil.ParseTaskSelection(s.viewDef.DefaultQuery, selectors, s.qql, s.rng, s.str)
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
		for _, tagOp := range tagOps {
			eqIdx := strings.Index(tagOp, "=")
			if err := t.SetTag(tagOp[:eqIdx], tagOp[eqIdx+1:]); err != nil {
				return err
			}
		}

	}
	for _, tagOp := range tagOps {
		eqIdx := strings.Index(tagOp, "=")
		tag, value := tagOp[:eqIdx], tagOp[eqIdx+1:]
		if value == "" {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Removed tag \"%s\" from", tag), list, confirmedSelection)
		} else {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Set tag \"%s\" to \"%s\" on", tag, value), list, confirmedSelection)
		}
	}
	return nil
}
