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
	rootCmd.PersistentPreRunE = cmdutil.Steps(
		cmdutil.EnsureTodoFileExits,
		cmdutil.EnsureDoneFileExists,
		cmdutil.RegisterMacros,
		cmdutil.SyncConflictProtection,
	)
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().String("config", "", "the config file to use") // This is just for the help message. Parsing happens in main.go
	rootCmd.PersistentFlags().StringP("file", "f", "", "the todo.txt file")
	di.Config().BindPFlag(config.TodoFile, rootCmd.PersistentFlags().Lookup("file"))
	rootCmd.PersistentFlags().BoolP("interactive", "i", true, "set to false to make list commands non-interactive")
	di.Config().BindPFlag(config.Interactive, rootCmd.PersistentFlags().Lookup("interactive"))
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	rootCmd.AddCommand(newOpenCommand().command())
	for _, def := range di.ViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}

	return rootCmd
}
func Execute(di *config.Di, args []string) {
	ctx := context.WithValue(context.Background(), cmdutil.DiKey, di)
	root := Root(di)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
