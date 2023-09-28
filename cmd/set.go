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

var setTagRegex = regexp.MustCompile("^[^[:space:]]+:[^[:space:]]+$")
var setProjectRegex = regexp.MustCompile(`^\+[^[:space:]]+$`)
var setContextRegex = regexp.MustCompile("^@[^[:space:]]+$")

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
	projectsOps, contextOps, tagOps, selectors := s.parseArgs(args)
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

	var projectAdded map[todotxt.Project][]*todotxt.Item = make(map[todotxt.Project][]*todotxt.Item)
	var contextAdded map[todotxt.Context][]*todotxt.Item
	for _, t := range confirmedSelection {
		err = s.applyTags(t, tagOps)
		if err != nil {
			return err
		}

		contextAdded, err = s.applyContexts(t, contextOps)
		if err != nil {
			return err
		}

		projectAdded, err = s.applyProjects(t, projectsOps)
		if err != nil {
			return err
		}
	}

	for key, value := range tagOps {
		cmdutil.PrintSuccessMessage(fmt.Sprintf("Set tag \"%s\" to \"%s\" on", key, value), list, confirmedSelection)
	}
	for _, context := range contextOps {
		if len(contextAdded[context]) > 0 {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Set context \"%s\" on", context), list, contextAdded[context])
		}
	}
	for _, project := range projectsOps {
		if len(projectAdded[project]) > 0 {
			cmdutil.PrintSuccessMessage(fmt.Sprintf("Set Project \"%s\" on", project), list, projectAdded[project])
		}
	}
	if len(tagOps)+len(projectAdded)+len(contextAdded) == 0 {
		fmt.Println("nothing to do")
	}
	return nil
}

func (s *setCommand) applyTags(t *todotxt.Item, tagOps map[string]string) error {
	for key, value := range tagOps {
		if err := t.SetTag(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *setCommand) applyContexts(t *todotxt.Item, contextOps []todotxt.Context) (map[todotxt.Context][]*todotxt.Item, error) {
	contextAdded := make(map[todotxt.Context][]*todotxt.Item)
	contexts := t.Contexts()
	for _, context := range contextOps {
		if slices.Contains(contexts, context) {
			continue
		}
		if err := t.EditDescription(fmt.Sprintf("%s %s", context, t.Description())); err != nil {
			return nil, err
		}
		contextAdded[context] = append(contextAdded[context], t)
	}
	return contextAdded, nil
}

func (s *setCommand) applyProjects(t *todotxt.Item, projectsOps []todotxt.Project) (map[todotxt.Project][]*todotxt.Item, error) {
	projectsAdded := make(map[todotxt.Project][]*todotxt.Item)
	projects := t.Projects()
	for _, project := range projectsOps {
		if slices.Contains(projects, project) {
			continue
		}
		if err := t.EditDescription(fmt.Sprintf("%s %s", project, t.Description())); err != nil {
			return nil, err
		}
		projectsAdded[project] = append(projectsAdded[project], t)
	}
	return projectsAdded, nil
}

func (s *setCommand) parseArgs(args []string) (projects []todotxt.Project, contexts []todotxt.Context, tags map[string]string, selectors []string) {
	tags = make(map[string]string)

	if idx := slices.Index(args, "on"); idx != -1 {
		selectors = args[idx+1:]
		args = slices.Delete(args, idx, len(args))
	}

	for _, arg := range args {
		switch {
		case setTagRegex.MatchString(arg):
			sepIdx := strings.Index(arg, ":")
			tags[arg[:sepIdx]] = arg[sepIdx+1:]
		case setProjectRegex.MatchString(arg):
			projects = append(projects, todotxt.Project(strings.TrimPrefix(arg, "+")))
		case setContextRegex.MatchString(arg):
			contexts = append(contexts, todotxt.Context(strings.TrimPrefix(arg, "@")))
		default:
			selectors = append(selectors, arg)
		}
	}
	return
}
