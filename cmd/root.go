package cmd

import (
	"context"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/spf13/cobra"
)

func Root(di *di.Container) (*cobra.Command, context.Context) {
	defaultView := di.Config().DefaultView
	rootCmd := newViewCommand(defaultView).command("")
	rootCmd.PersistentFlags().String("config", "", "the config file to use") // This is just for the help message. Parsing happens in main.go
	rootCmd.PersistentFlags().StringP("file", "f", "", "the todo.txt file")
	rootCmd.PersistentPreRunE = cmdutil.Steps(
		cmdutil.ConfigOverrides,
		cmdutil.EnsureTodoFileExits,
		cmdutil.EnsureDoneFileExists,
		cmdutil.RegisterMacros,
		cmdutil.SyncConflictProtection,
	)
	rootCmd.SilenceUsage = true

	rootCmd.AddGroup(&cobra.Group{
		ID:    "view",
		Title: "Views",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "global-cmd",
		Title: "Global Command",
	})

	rootCmd.AddCommand(newOpenCommand().command())
	for name, def := range di.Config().Views {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command(name))
	}

	return rootCmd, context.WithValue(context.Background(), cmdutil.DiKey, di)
}
