package cmd

import (
	"fmt"
	"slices"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
)

type viewCommand struct {
	def          di.ViewDef
	projection   []string
	sortOrder    []string
	qqlSearch    []string
	rngSearch    []string
	stringSearch []string
	json         bool
	interactive  bool
}

func newViewCommand(def di.ViewDef) *viewCommand {
	cmd := viewCommand{
		def: def,
	}

	return &cmd
}

func (v *viewCommand) command(name string) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:     AppName + " " + name + " [selectors...]",
		Short:   "Lists all tasks that match the view query",
		GroupID: "view",
		PreRunE: cmdutil.Steps(cmdutil.LoadList),
		RunE:    v.list,
	}
	listCmd.AddGroup(&cobra.Group{
		ID:    "view-cmd",
		Title: "View Commands",
	})

	listCmd.Flags().StringSliceVarP(&v.projection, "projection", "p", v.def.Projection, "TODO")
	listCmd.Flags().StringSliceVarP(&v.sortOrder, "sort", "s", v.def.Sort, "TODO")
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
	di := cmd.Context().Value(cmdutil.DiKey).(*di.Container)
	repo := di.TodoTxtRepo()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	query, err := cmdutil.ParseTaskSelection(v.def.Query, args, v.qqlSearch, v.rngSearch, v.stringSearch)
	if err != nil {
		return fmt.Errorf("invalid query specified: %w", err)
	}

	sortCompiler := qsort.Compiler{
		TagTypes:        di.Config().TagTypes(),
		ScoreCalculator: di.QuestScoreCalculator(),
	}
	sortFunc, err := sortCompiler.CompileSortFunc(v.sortOrder)
	if err != nil {
		return err
	}

	projector := di.Projector(cmd)
	if err := projector.Verify(v.projection, list); err != nil {
		return fmt.Errorf("invalid projection: %w", err)
	}

	selection := query.Filter(list)
	slices.SortStableFunc(selection, sortFunc)

	if v.json {
		return todotxt.DefaultJsonEncoder.Encode(cmd.OutOrStdout(), list, selection)
	}

	return view.NewList(repo, projector, v.projection, query, sortFunc, v.interactive).Run()
}
