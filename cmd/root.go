package cmd

import (
	"os"

	"github.com/Fabian-G/quest/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	viewCommand := newViewCommand(config.ListViewDef)
	rootCmd.AddCommand(viewCommand.command())
	for _, def := range config.GetViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}
}
