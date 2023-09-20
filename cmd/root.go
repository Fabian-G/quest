package cmd

import (
	"context"
	"os"

	"github.com/Fabian-G/quest/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute(di *config.Di) {
	defaultView := di.DefaultViewDef()
	rootCmd := newViewCommand(defaultView).command()
	rootCmd.PersistentPreRunE = steps(ensureTodoFileExits, registerMacros)
	rootCmd.PersistentFlags().StringP("file", "f", "", "overrides the todo txt file location")
	viper.BindPFlag(config.TodoFile, rootCmd.PersistentFlags().Lookup("file"))
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	for _, def := range di.ViewDefs() {
		viewCommand := newViewCommand(def)
		rootCmd.AddCommand(viewCommand.command())
	}

	ctx := context.WithValue(context.Background(), diKey, di)
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
