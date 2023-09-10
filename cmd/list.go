package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	interactive bool
	projection  []string
)
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "TODO",
	GroupID: "query",
	Long:    `TODO`,
	Example: "TODO",
	RunE:    list,
}

func init() {
	listCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "TODO")
	listCmd.Flags().StringSliceVarP(&projection, "projection", "p", []string{"idx", "description"}, "TODO")

}

func list(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(listKey).(*todotxt.List)
	selection := cmd.Context().Value(selectionKey).([]*todotxt.Item)

	listView, err := view.NewList(list, selection, projection, interactive)
	if err != nil {
		return fmt.Errorf("could not create list view: %w", err)
	}
	programme := tea.NewProgram(listView)
	if _, err := programme.Run(); err != nil {
		return err
	}
	return nil
}
