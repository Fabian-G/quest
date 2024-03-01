package cmd

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
)

var setTagRegex = regexp.MustCompile("^[^[:space:]]+:[^[:space:]]+$")
var setProjectRegex = regexp.MustCompile(`^\+[^[:space:]]+$`)
var setContextRegex = regexp.MustCompile("^@[^[:space:]]+$")

type setCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newSetCommand(def di.ViewDef) *setCommand {
	cmd := setCommand{
		viewDef: def,
	}

	return &cmd
}

func (s *setCommand) command() *cobra.Command {
	var setCommand = &cobra.Command{
		Use:      "set [attributes...] on [selectors...]",
		Args:     cobra.MinimumNArgs(1),
		Short:    "Sets the given attributes (+project, @context, key:value) on the matching tasks.",
		Example:  "quest set due:today @errands on 4,5",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     s.set,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(setCommand, &s.qql, &s.rng, &s.str, &s.all)
	return setCommand
}

func (s *setCommand) set(cmd *cobra.Command, args []string) error {
	projectsOps, contextOps, tagOps, selectors := s.parseArgs(args)
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)

	selector, err := cmdutil.ParseTaskSelection(s.viewDef.Query, selectors, s.qql, s.rng, s.str)
	if err != nil {
		return err
	}

	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !s.all {
		confirmedSelection, err = view.NewSelection(selection).Run()
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Println("no matches")
		return nil
	}

	var projectAdded map[todotxt.Project][]*todotxt.Item = make(map[todotxt.Project][]*todotxt.Item)
	var contextAdded map[todotxt.Context][]*todotxt.Item = make(map[todotxt.Context][]*todotxt.Item)
	for _, t := range confirmedSelection {
		err = s.applyTags(t, tagOps)
		if err != nil {
			return err
		}

		addedContexts, err := s.applyContexts(t, contextOps)
		if err != nil {
			return err
		}
		contextAdded = mergeChanges(contextAdded, t, addedContexts)

		addedProjects, err := s.applyProjects(t, projectsOps)
		if err != nil {
			return err
		}
		projectAdded = mergeChanges(projectAdded, t, addedProjects)
	}

	for key, value := range tagOps {
		view.NewSuccessMessage(fmt.Sprintf("Set tag \"%s\" to \"%s\" on", key, value), list, confirmedSelection).Run()
	}
	for _, context := range contextOps {
		if len(contextAdded[context]) > 0 {
			view.NewSuccessMessage(fmt.Sprintf("Set context \"%s\" on", context), list, contextAdded[context]).Run()
		}
	}
	for _, project := range projectsOps {
		if len(projectAdded[project]) > 0 {
			view.NewSuccessMessage(fmt.Sprintf("Set Project \"%s\" on", project), list, projectAdded[project]).Run()
		}
	}
	if len(tagOps)+len(projectAdded)+len(contextAdded) == 0 {
		fmt.Println("nothing to do")
	}
	return nil
}

func mergeChanges[T comparable](current map[T][]*todotxt.Item, item *todotxt.Item, additions []T) map[T][]*todotxt.Item {
	for _, a := range additions {
		current[a] = append(current[a], item)
	}
	return current
}

func (s *setCommand) applyTags(t *todotxt.Item, tagOps map[string]string) error {
	for key, value := range tagOps {
		if err := t.SetTag(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *setCommand) applyContexts(t *todotxt.Item, contextOps []todotxt.Context) ([]todotxt.Context, error) {
	contextAdded := make([]todotxt.Context, 0)
	contexts := t.Contexts()
	for _, context := range contextOps {
		if slices.Contains(contexts, context) {
			continue
		}
		if err := t.EditDescription(fmt.Sprintf("%s %s", t.Description(), context)); err != nil {
			return nil, err
		}
		contextAdded = append(contextAdded, context)
	}
	return contextAdded, nil
}

func (s *setCommand) applyProjects(t *todotxt.Item, projectsOps []todotxt.Project) ([]todotxt.Project, error) {
	projectsAdded := make([]todotxt.Project, 0)
	projects := t.Projects()
	for _, project := range projectsOps {
		if slices.Contains(projects, project) {
			continue
		}
		if err := t.EditDescription(fmt.Sprintf("%s %s", t.Description(), project)); err != nil {
			return nil, err
		}
		projectsAdded = append(projectsAdded, project)
	}
	return projectsAdded, nil
}

func (s *setCommand) parseArgs(args []string) (projects []todotxt.Project, contexts []todotxt.Context, tags map[string]string, selectors []string) {
	tags = make(map[string]string)

	if idx := slices.Index(args, "on"); idx != -1 {
		selectors = args[idx+1:]
		args = args[:idx]
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
