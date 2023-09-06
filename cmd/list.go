package cmd

import (
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	query       string
	interactive bool
)
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list shows the items in your todo.txt optionally filtered by query",
	GroupID: "query",
	Long: `list is the basic command with which you can query your todo.txt file.
With the -s option you can specify a filter, which can either be a FOL formula, a range expression or a string search.
		`,
	Example: "list -s 1-2,4",
	RunE:    list,
}

func init() {
	listCmd.Flags().StringVarP(&query, "select", "s", "", "can be a FOL query, range expression or a string search")
	listCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "whether or not quest should be interactive")
}

func list(cmd *cobra.Command, args []string) error {
	di := config.Di{}
	repo := di.TodoTxtRepo()
	list, err := repo.Read()
	if err != nil {
		return err
	}
	listView, err := view.NewList(list, query, interactive)
	if err != nil {
		return err
	}
	programme := tea.NewProgram(listView)
	if _, err := programme.Run(); err != nil {
		return err
	}
	return nil
}
