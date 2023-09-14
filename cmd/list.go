package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type outputMode = string

const (
	jsonOutput        outputMode = "json"
	interactiveOutput outputMode = "interactive"
	listOutputMode    outputMode = "list"
)

type viewDef struct {
	name              string
	defaultSelection  string
	defaultProjection string
	defaultSortOrder  string
	defaultOutputMode outputMode
	defaultClean      []string
}

func listCmd(v viewDef) *cobra.Command {
	var (
		projection []string
		clean      []string
		sortOrder  string
		output     outputMode
	)

	list := func(cmd *cobra.Command, args []string) error {
		list := cmd.Context().Value(listKey).(*todotxt.List)
		userQuery := cmd.Context().Value(queryKey).(query.Func)
		configQuery, err := query.Compile(v.defaultSelection, query.QQL)
		if err != nil {
			return fmt.Errorf("config file contains invalid query for view %s: %w", v.name, err)
		}
		selection := query.And(userQuery, configQuery).Filter(list)
		if len(selection) == 0 {
			fmt.Println("no matches")
			return nil
		}

		sortFunc, err := query.SortFunc(sortOrder)
		if err != nil {
			return err
		}
		slices.SortFunc(selection, sortFunc)

		listView, err := view.NewList(list, selection, projection)
		if err != nil {
			return fmt.Errorf("could not create list view: %w", err)
		}
		listView.Interactive = output == interactiveOutput
		cleanProjects, cleanContexts, cleanTags := cleanAttributes(list, clean)
		listView.CleanProjects = cleanProjects
		listView.CleanContexts = cleanContexts
		listView.CleanTags = cleanTags
		programme := tea.NewProgram(listView)
		if _, err := programme.Run(); err != nil {
			return err
		}
		return nil
	}

	var listCmd = &cobra.Command{
		Use:     v.name,
		Short:   "TODO",
		GroupID: "query",
		Long:    `TODO `,
		Example: "TODO",
		RunE:    list,
	}

	listCmd.Flags().StringVarP(&output, "output", "o", v.defaultOutputMode, "TODO")
	listCmd.Flags().StringSliceVarP(&projection, "projection", "p", strings.Split(v.defaultProjection, ","), "TODO")
	listCmd.Flags().StringVarP(&sortOrder, "sort", "s", v.defaultSortOrder, "TODO")
	listCmd.Flags().StringSliceVarP(&clean, "clean", "c", v.defaultClean, "TODO")

	return listCmd
}

func cleanAttributes(list *todotxt.List, clean []string) (proj []todotxt.Project, ctx []todotxt.Context, tags []string) {
	for _, c := range clean {
		c := strings.TrimSpace(c)
		switch {
		case c == "+ALL":
			proj = append(proj, list.AllProjects()...)
		case c == "@ALL":
			ctx = append(ctx, list.AllContexts()...)
		case c == "ALL":
			tags = append(tags, list.AllTags()...)
		case strings.HasPrefix(c, "@"):
			ctx = append(ctx, todotxt.Context(c[1:]))
		case strings.HasPrefix(c, "+"):
			proj = append(proj, todotxt.Project(c[1:]))
		case len(c) == 0:
			continue
		default:
			tags = append(tags, c)
		}
	}
	return
}
