package cmd

import (
	"fmt"
	"slices"

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
	sortOrder    []string
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

	listCmd.Flags().StringSliceVarP(&v.projection, "projection", "p", v.def.DefaultProjection, "TODO")
	listCmd.Flags().StringSliceVarP(&v.sortOrder, "sort", "s", v.def.DefaultSortOrder, "TODO")
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
	listCmd.AddCommand(newSetCommand(v.def).command())
	listCmd.AddCommand(newUnsetCommand(v.def).command())

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

	sortCompiler := qsort.Compiler{
		TagTypes:        di.TagTypes(),
		ScoreCalculator: di.QuestScoreCalculator(),
	}
	sortFunc, err := sortCompiler.CompileSortFunc(v.sortOrder)
	if err != nil {
		return err
	}

	projector := qprojection.Projector{
		Clean:     v.clean,
		ScoreCalc: di.QuestScoreCalculator(),
	}
	if err := projector.Verify(v.projection, list); err != nil {
		return fmt.Errorf("invalid projection: %w", err)
	}

	selection := query.Filter(list)
	slices.SortStableFunc(selection, sortFunc)

	if v.json {
		return todotxt.DefaultJsonEncoder.Encode(cmd.OutOrStdout(), selection)
	}

	listView, err := view.NewList(projector, v.interactive)
	if err != nil {
		return fmt.Errorf("could not create list view: %w", err)
	}
	model, _ := listView.Update(view.RefreshListMsg{
		List:       list,
		Selection:  selection,
		Projection: v.projection,
	})
	switch {
	case v.interactive:
		programme := tea.NewProgram(model, tea.WithOutput(cmd.OutOrStdout()))
		end := startAutoUpdate(repo, programme, v.projection, query, sortFunc)
		defer end()
		if _, err := programme.Run(); err != nil {
			return err
		}
	default:
		fmt.Fprint(cmd.OutOrStdout(), model.View())
	}
	return nil
}

func startAutoUpdate(repo *todotxt.Repo, prog *tea.Program, projection []string, query qselect.Func, sort func(*todotxt.Item, *todotxt.Item) int) func() {
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
			slices.SortStableFunc(selection, sort)
			prog.Send(view.RefreshListMsg{
				List:       newList,
				Selection:  selection,
				Projection: projection,
			})
		}
	}()
	return end
}
