package cmd

import (
	"context"
	"os"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/spf13/cobra"
)

func Root(di *config.Di) *cobra.Command {
	defaultView := di.DefaultViewDef()
	rootCmd := newViewCommand(defaultView).command()
	rootCmd.PersistentPreRunE = cmdutil.Steps(cmdutil.EnsureTodoFileExits, cmdutil.RegisterMacros)
	rootCmd.PersistentFlags().StringP("file", "f", "", "the todo.txt file")
	di.Config().BindPFlag(config.TodoFile, rootCmd.PersistentFlags().Lookup("file"))
	rootCmd.PersistentFlags().BoolP("interactive", "i", true, "set to false to make list commands non-interactive")
	di.Config().BindPFlag(config.Interactive, rootCmd.PersistentFlags().Lookup("interactive"))
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	for _, def := range di.ViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}

	return rootCmd
}
func Execute(di *config.Di) {
	ctx := context.WithValue(context.Background(), cmdutil.DiKey, di)
	err := Root(di).ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
