package cmd

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

var unsetTagRegex = regexp.MustCompile("^[^[:space:]]+$")

type unsetCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newUnsetCommand(def config.ViewDef) *unsetCommand {
	cmd := unsetCommand{
		viewDef: def,
	}

	return &cmd
}

func (u *unsetCommand) command() *cobra.Command {
	var unsetCommand = &cobra.Command{
		Use:      "unset",
		Args:     cobra.MinimumNArgs(1),
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     u.unset,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	unsetCommand.Flags().BoolVarP(&u.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(unsetCommand, &u.qql, &u.rng, &u.str)
	return unsetCommand
}

func (u *unsetCommand) unset(cmd *cobra.Command, args []string) error {
	projectsOps, contextOps, tagOps, selectors := u.parseArgs(args)
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)

	selector, err := cmdutil.ParseTaskSelection(u.viewDef.DefaultQuery, selectors, u.qql, u.rng, u.str)
	if err != nil {
		return err
	}

	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !u.all {
		confirmedSelection, err = cmdutil.ConfirmSelection(selection)
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}

	var projectsRemoved map[todotxt.Project][]*todotxt.Item = make(map[todotxt.Project][]*todotxt.Item)
	var contextsRemoved map[todotxt.Context][]*todotxt.Item = make(map[todotxt.Context][]*todotxt.Item)
	var tagsRemoved map[string][]*todotxt.Item = make(map[string][]*todotxt.Item)
	for _, t := range confirmedSelection {
		tags := t.Tags()
		for _, tag := range tagOps {
			if _, ok := tags[tag]; !ok {
				continue
			}
			if err := t.SetTag(tag, ""); err != nil {
				return err
			}
			tagsRemoved[tag] = append(tagsRemoved[tag], t)
		}

		contexts := t.Contexts()
		for _, context := range contextOps {
			if !slices.Contains(contexts, context) {
				continue
			}
			if err := t.EditDescription(t.CleanDescription(nil, []todotxt.Context{context}, nil)); err != nil {
				return err
			}
			contextsRemoved[context] = append(contextsRemoved[context], t)
		}

		projects := t.Projects()
		for _, project := range projectsOps {
			if !slices.Contains(projects, project) {
				continue
			}
			if err := t.EditDescription(t.CleanDescription([]todotxt.Project{project}, nil, nil)); err != nil {
				return err
			}
			projectsRemoved[project] = append(projectsRemoved[project], t)
		}
	}

	for _, tag := range tagOps {
		if len(tagsRemoved[tag]) > 0 {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Removed tag \"%s\" from", tag), list, tagsRemoved[tag])
		}
	}
	for _, context := range contextOps {
		if len(contextsRemoved[context]) > 0 {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Removed context \"%s\" from", context), list, contextsRemoved[context])
		}
	}
	for _, project := range projectsOps {
		if len(projectsRemoved[project]) > 0 {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Removed Project \"%s\" from", project), list, projectsRemoved[project])
		}
	}
	if len(tagsRemoved)+len(projectsRemoved)+len(contextsRemoved) == 0 {
		fmt.Println("nothing to do")
	}
	return nil
}

func (u *unsetCommand) parseArgs(args []string) (projects []todotxt.Project, contexts []todotxt.Context, tags []string, selectors []string) {
	if idx := slices.Index(args, "on"); idx != -1 {
		selectors = args[idx+1:]
		args = slices.Delete(args, idx, len(args))
	}

	for _, arg := range args {
		switch {
		case setProjectRegex.MatchString(arg):
			projects = append(projects, todotxt.Project(strings.TrimPrefix(arg, "+")))
		case setContextRegex.MatchString(arg):
			contexts = append(contexts, todotxt.Context(strings.TrimPrefix(arg, "@")))
		case unsetTagRegex.MatchString(arg):
			tags = append(tags, arg)
		default:
			selectors = append(selectors, arg)
		}
	}
	return
}
