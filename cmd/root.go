package cmd

import (
	"context"

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
	di.Config().BindPFlag(config.TodoFileKey, rootCmd.PersistentFlags().Lookup("file"))
	rootCmd.AddGroup(&cobra.Group{
		ID:    "view",
		Title: "Views",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "global-cmd",
		Title: "Global Command",
	})

	rootCmd.AddCommand(newOpenCommand().command())
	for _, def := range di.ViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}

	return rootCmd
}
func Execute(di *config.Di, args []string) error {
	ctx := context.WithValue(context.Background(), cmdutil.DiKey, di)
	root := Root(di)
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}
