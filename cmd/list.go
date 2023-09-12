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

func listCmd(name, defaultSelection, defaultProjection, defaultSortOrder string, defaultOutputMode outputMode) *cobra.Command {
	var (
		projection []string
		sortOrder  string
		output     outputMode
	)

	list := func(cmd *cobra.Command, args []string) error {
		list := cmd.Context().Value(listKey).(*todotxt.List)
		userQuery := cmd.Context().Value(queryKey).(query.Func)
		configQuery, err := query.Compile(defaultSelection, query.QQL)
		if err != nil {
			return fmt.Errorf("config file contains invalid query for view %s: %w", name, err)
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

		listView, err := view.NewList(list, selection, projection, output == interactiveOutput)
		if err != nil {
			return fmt.Errorf("could not create list view: %w", err)
		}
		programme := tea.NewProgram(listView)
		if _, err := programme.Run(); err != nil {
			return err
		}
		return nil
	}

	var listCmd = &cobra.Command{
		Use:     name,
		Short:   "TODO",
		GroupID: "query",
		Long:    `TODO `,
		Example: "TODO",
		RunE:    list,
	}

	listCmd.Flags().StringVarP(&output, "output", "o", defaultOutputMode, "TODO")
	listCmd.Flags().StringSliceVarP(&projection, "projection", "p", strings.Split(defaultProjection, ","), "TODO")
	listCmd.Flags().StringVarP(&sortOrder, "sort", "s", defaultSortOrder, "TODO")

	return listCmd
}
