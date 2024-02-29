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

var unsetTagRegex = regexp.MustCompile("^[^[:space:]]+$")

type unsetCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newUnsetCommand(def di.ViewDef) *unsetCommand {
	cmd := unsetCommand{
		viewDef: def,
	}

	return &cmd
}

func (u *unsetCommand) command() *cobra.Command {
	var unsetCommand = &cobra.Command{
		Use:      "unset [attributes...] on [selectors...]",
		Args:     cobra.MinimumNArgs(1),
		Short:    "Removes the given attributes (+project, @context, tag) from all matching items.",
		Example:  "quest unset +inbox on 4",
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

	selector, err := cmdutil.ParseTaskSelection(u.viewDef.Query, selectors, u.qql, u.rng, u.str)
	if err != nil {
		return err
	}

	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !u.all {
		confirmedSelection, err = view.NewSelection(selection).Run()
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Println("no matches")
		return nil
	}

	var projectsRemoved map[todotxt.Project][]*todotxt.Item = make(map[todotxt.Project][]*todotxt.Item)
	var contextsRemoved map[todotxt.Context][]*todotxt.Item = make(map[todotxt.Context][]*todotxt.Item)
	var tagsRemoved map[string][]*todotxt.Item = make(map[string][]*todotxt.Item)
	for _, t := range confirmedSelection {
		removedTags, err := u.applyTags(t, tagOps)
		if err != nil {
			return err
		}
		tagsRemoved = mergeChanges(tagsRemoved, t, removedTags)

		removedContexts, err := u.applyContexts(t, contextOps)
		if err != nil {
			return err
		}
		contextsRemoved = mergeChanges(contextsRemoved, t, removedContexts)

		removedProjects, err := u.applyProjects(t, projectsOps)
		if err != nil {
			return err
		}
		projectsRemoved = mergeChanges(projectsRemoved, t, removedProjects)
	}

	u.printTagChanges(list, tagOps, tagsRemoved)
	u.printProjectChanges(list, projectsOps, projectsRemoved)
	u.printContextChanges(list, contextOps, contextsRemoved)
	if len(tagsRemoved)+len(projectsRemoved)+len(contextsRemoved) == 0 {
		fmt.Println("nothing to do")
	}
	return nil
}

func (u *unsetCommand) printTagChanges(list *todotxt.List, tagOps []string, removals map[string][]*todotxt.Item) {
	for _, tag := range tagOps {
		if len(removals[tag]) > 0 {
			view.NewSuccessMessage(fmt.Sprintf("Removed tag \"%s\" from", tag), list, removals[tag]).Run()
		}
	}
}

func (u *unsetCommand) printContextChanges(list *todotxt.List, contextOps []todotxt.Context, removals map[todotxt.Context][]*todotxt.Item) {
	for _, context := range contextOps {
		if len(removals[context]) > 0 {
			view.NewSuccessMessage(fmt.Sprintf("Removed context \"%s\" from", context), list, removals[context]).Run()
		}
	}
}

func (u *unsetCommand) printProjectChanges(list *todotxt.List, projectOps []todotxt.Project, removals map[todotxt.Project][]*todotxt.Item) {
	for _, project := range projectOps {
		if len(removals[project]) > 0 {
			view.NewSuccessMessage(fmt.Sprintf("Removed Project \"%s\" from", project), list, removals[project]).Run()
		}
	}
}

func (u *unsetCommand) applyTags(t *todotxt.Item, tagOps []string) ([]string, error) {
	tagsRemoved := make([]string, 0)
	tags := t.Tags()
	for _, tag := range tagOps {
		if _, ok := tags[tag]; !ok {
			continue
		}
		if err := t.SetTag(tag, ""); err != nil {
			return nil, err
		}
		tagsRemoved = append(tagsRemoved, tag)
	}
	return tagsRemoved, nil
}

func (u *unsetCommand) applyProjects(t *todotxt.Item, projectOps []todotxt.Project) ([]todotxt.Project, error) {
	projectsRemoved := make([]todotxt.Project, 0)
	projects := t.Projects()
	for _, project := range projectOps {
		if !slices.Contains(projects, project) {
			continue
		}
		if err := t.EditDescription(t.CleanDescription([]todotxt.Project{project}, nil, nil)); err != nil {
			return nil, err
		}
		projectsRemoved = append(projectsRemoved, project)
	}
	return projectsRemoved, nil
}

func (u *unsetCommand) applyContexts(t *todotxt.Item, contextOps []todotxt.Context) ([]todotxt.Context, error) {
	contextsRemoved := make([]todotxt.Context, 0)
	contexts := t.Contexts()
	for _, context := range contextOps {
		if !slices.Contains(contexts, context) {
			continue
		}
		if err := t.EditDescription(t.CleanDescription(nil, []todotxt.Context{context}, nil)); err != nil {
			return nil, err
		}
		contextsRemoved = append(contextsRemoved, context)
	}
	return contextsRemoved, nil
}

func (u *unsetCommand) parseArgs(args []string) (projects []todotxt.Project, contexts []todotxt.Context, tags []string, selectors []string) {
	if idx := slices.Index(args, "on"); idx != -1 {
		selectors = args[idx+1:]
		args = args[:idx]
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
