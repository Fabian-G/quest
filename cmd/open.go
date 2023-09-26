package cmd

import (
	"os"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/spf13/cobra"
)

type openCommand struct {
}

func newOpenCommand() *openCommand {
	cmd := openCommand{}

	return &cmd
}

func (o *openCommand) command() *cobra.Command {
	var openCommand = &cobra.Command{
		Use:   "open",
		Short: "Opens your todo file in the editor",
		Long: `Open opens your todo file in your editor.
Note that this is just as if you would open your todo.txt file in your editor directly, 
but with the extra benefit that your changes will be validated afterwards.
However, special features like recurrence or tag expansion will not be triggered by your changes.`,
		Example: "quest open",
		GroupID: "global-cmd",
		RunE:    o.open,
	}
	return openCommand
}

func (o *openCommand) open(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(cmdutil.DiKey).(*config.Di)
	cfg := di.Config()
	repo := di.TodoTxtRepo()
	editor := di.Editor()

	for {
		if err := editor.Edit(cfg.GetString(config.TodoFileKey)); err != nil {
			return err
		}
		todoFile, err := os.Open(cfg.GetString(config.TodoFileKey))
		if err != nil {
			return err
		}
		defer todoFile.Close()
		if _, err = repo.Read(); err == nil {
			return nil
		}
		if !cmdutil.AskRetry(cmd, err) {
			return err
		}
	}
}
