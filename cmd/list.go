package cmd

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type viewCommand struct {
	def          config.ViewDef
	projection   []string
	clean        []string
	sortOrder    string
	output       config.OutputMode
	qqlSearch    []string
	rngSearch    []string
	stringSearch []string
}

func newViewCommand(def config.ViewDef) *viewCommand {
	cmd := viewCommand{
		def: def,
	}

	return &cmd
}

func (v *viewCommand) command() *cobra.Command {
	var listCmd = &cobra.Command{
		Use:     v.def.Name,
		Short:   "TODO",
		GroupID: "query",
		Long:    `TODO `,
		Example: "TODO",
		PreRunE: steps(initDI, loadList),
		RunE:    v.list,
	}

	listCmd.Flags().StringVarP(&v.output, "output", "o", v.def.DefaultOutputMode, "TODO")
	listCmd.Flags().StringSliceVarP(&v.projection, "projection", "p", strings.Split(v.def.DefaultProjection, ","), "TODO")
	listCmd.Flags().StringVarP(&v.sortOrder, "sort", "s", v.def.DefaultSortOrder, "TODO")
	listCmd.Flags().StringSliceVarP(&v.clean, "clean", "c", v.def.DefaultClean, "TODO")
	listCmd.Flags().StringArrayVarP(&v.qqlSearch, "qql", "q", nil, "TODO")
	listCmd.Flags().StringArrayVarP(&v.rngSearch, "range", "r", nil, "TODO")
	listCmd.Flags().StringArrayVarP(&v.stringSearch, "word", "w", nil, "TODO")

	listCmd.AddCommand(newAddCommand(v.def).command())

	return listCmd
}

func (v *viewCommand) list(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(listKey).(*todotxt.List)
	query, err := parseTaskSelection(v.def.DefaultQuery, args, v.qqlSearch, v.rngSearch, v.stringSearch)
	if err != nil {
		return fmt.Errorf("invalid query specified: %w", err)
	}
	selection := query.Filter(list)
	if len(selection) == 0 {
		fmt.Println("no matches")
		return nil
	}

	sortFunc, err := qsort.CompileSortFunc(v.sortOrder, config.TagTypes())
	if err != nil {
		return err
	}
	slices.SortFunc(selection, sortFunc)

	cleanProjects, cleanContexts, cleanTags := cleanAttributes(list, v.clean)
	projectionCfg := qprojection.Config{
		ColumnNames:   v.projection,
		List:          list,
		CleanTags:     cleanTags,
		CleanProjects: cleanProjects,
		CleanContexts: cleanContexts,
	}
	listView, err := view.NewList(list, selection, v.output == config.InteractiveOutput, projectionCfg)
	if err != nil {
		return fmt.Errorf("could not create list view: %w", err)
	}
	switch v.output {
	case config.InteractiveOutput:
		programme := tea.NewProgram(listView)
		if _, err := programme.Run(); err != nil {
			return err
		}
	case config.JsonOutput:
		log.Fatal("Not implemented")
	case config.ListOutput:
		l, _ := listView.Update(view.RefreshList())
		fmt.Println(l.View())
	default:
		return fmt.Errorf("unknown output mode: %s\nAvailable modes are %v", v.output, []string{config.InteractiveOutput, config.JsonOutput, config.ListOutput})
	}
	return nil
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

func parseTaskSelection(defaultQuery string, guess, qqlSearch, rngSearch, stringSearch []string) (qselect.Func, error) {
	selectors := make([]qselect.Func, 0)
	q, err := qselect.CompileQQL(defaultQuery)
	if err != nil {
		return nil, fmt.Errorf("config file contains invalid query: %w", err)
	}
	selectors = append(selectors, q)
	for _, arg := range guess {
		q, err := qselect.CompileQuery(arg)
		if err != nil {
			return nil, fmt.Errorf("could not compile query %s. Try using -q,-r or -s explicitly instead of positional args: %w", arg, err)
		}
		selectors = append(selectors, q)
	}
	for _, f := range qqlSearch {
		q, err := qselect.CompileQQL(f)
		if err != nil {
			return nil, fmt.Errorf("could not compile FOL query %s: %w", f, err)
		}
		selectors = append(selectors, q)
	}
	for _, r := range rngSearch {
		q, err := qselect.CompileRange(r)
		if err != nil {
			return nil, fmt.Errorf("could not compile range query %s: %w", r, err)
		}
		selectors = append(selectors, q)
	}
	for _, s := range stringSearch {
		q, err := qselect.CompileWordSearch(s)
		if err != nil {
			return nil, fmt.Errorf("could not compile string search query %s: %w", s, err)
		}
		selectors = append(selectors, q)
	}

	return qselect.And(selectors...), nil
}
