package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Fabian-G/quest/cmd/cmdutil"
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
	qqlSearch    []string
	rngSearch    []string
	stringSearch []string
	json         bool
	interactive  bool
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
		GroupID: "view",
		Long:    `TODO `,
		Example: "TODO",
		PreRunE: cmdutil.Steps(cmdutil.LoadList),
		RunE:    v.list,
	}
	listCmd.AddGroup(&cobra.Group{
		ID:    "view-cmd",
		Title: "View Commands",
	})

	listCmd.Flags().StringSliceVarP(&v.projection, "projection", "p", strings.Split(v.def.DefaultProjection, ","), "TODO")
	listCmd.Flags().StringVarP(&v.sortOrder, "sort", "s", v.def.DefaultSortOrder, "TODO")
	listCmd.Flags().StringSliceVarP(&v.clean, "clean", "c", v.def.DefaultClean, "TODO")
	listCmd.Flags().BoolVar(&v.json, "json", false, "TODO")
	listCmd.Flags().BoolVarP(&v.interactive, "interactive", "i", v.def.Interactive, "set to false to make the list non-interactive")
	cmdutil.RegisterSelectionFlags(listCmd, &v.qqlSearch, &v.rngSearch, &v.stringSearch)

	listCmd.AddCommand(newAddCommand(v.def).command())
	listCmd.AddCommand(newCompleteCommand(v.def).command())
	listCmd.AddCommand(newRemoveCommand(v.def).command())
	listCmd.AddCommand(newPrioritizeCommand(v.def).command())
	listCmd.AddCommand(newEditCommand(v.def).command())
	listCmd.AddCommand(newArchiveCommand(v.def).command())

	return listCmd
}

func (v *viewCommand) list(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(cmdutil.DiKey).(*config.Di)
	repo := di.TodoTxtRepo()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	query, err := cmdutil.ParseTaskSelection(v.def.DefaultQuery, args, v.qqlSearch, v.rngSearch, v.stringSearch)
	if err != nil {
		return fmt.Errorf("invalid query specified: %w", err)
	}

	sortFunc, err := qsort.CompileSortFunc(v.sortOrder, di.TagTypes())
	if err != nil {
		return err
	}

	cleanProjects, cleanContexts, cleanTags := cleanAttributes(list, v.clean)
	projectionCfg := qprojection.Config{
		ColumnNames:   v.projection,
		List:          list,
		CleanTags:     cleanTags,
		CleanProjects: cleanProjects,
		CleanContexts: cleanContexts,
	}

	selection := query.Filter(list)
	slices.SortStableFunc(selection, sortFunc)

	if v.json {
		return todotxt.DefaultJsonEncoder.Encode(cmd.OutOrStdout(), selection)
	}

	if len(selection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}
	listView, err := view.NewList(list, selection, v.interactive, projectionCfg)
	if err != nil {
		return fmt.Errorf("could not create list view: %w", err)
	}
	switch {
	case v.interactive || len(selection) == 1:
		programme := tea.NewProgram(listView, tea.WithOutput(cmd.OutOrStdout()))
		end := startAutoUpdate(repo, programme, projectionCfg, query, sortFunc)
		defer end()
		if _, err := programme.Run(); err != nil {
			return err
		}
	default:
		l, _ := listView.Update(view.RefreshListMsg{
			List:       list,
			Selection:  selection,
			Projection: projectionCfg,
		})
		fmt.Fprintln(cmd.OutOrStdout(), l.View())
	}
	return nil
}

func startAutoUpdate(repo *todotxt.Repo, prog *tea.Program, projectionCfg qprojection.Config, query qselect.Func, sort func(*todotxt.Item, *todotxt.Item) int) func() {
	news, end, err := repo.Watch()
	if err != nil {
		// Watching is not possible for whatever reason, but we ignore it to not interrupt the user
		return func() {}
	}
	go func() {
		for update := range news {
			newList, err := update()
			if err != nil {
				continue
			}
			selection := query.Filter(newList)
			projectionCfg.List = newList
			slices.SortStableFunc(selection, sort)
			prog.Send(view.RefreshListMsg{
				List:       newList,
				Selection:  selection,
				Projection: projectionCfg,
			})
		}
	}()
	return end
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
