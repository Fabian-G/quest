package cmd

import (
	"os"

	"github.com/Fabian-G/quest/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd *cobra.Command

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd = newViewCommand(config.DefaultViewDef()).command()
	rootCmd.PersistentFlags().StringP("file", "f", "", "overrides the todo txt file location")
	viper.BindPFlag(config.TodoFile, rootCmd.PersistentFlags().Lookup("file"))
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	for _, def := range config.GetViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}
}
