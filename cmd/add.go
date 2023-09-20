package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type addCommand struct {
	def config.AddDef
}

func newAddCommand(def config.ViewDef) *addCommand {
	cmd := addCommand{
		def: def.Add,
	}

	return &cmd
}

func (a *addCommand) command() *cobra.Command {
	var addCmd = &cobra.Command{
		Use:      "add",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     a.add,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}

	return addCmd
}

func (a *addCommand) add(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	description := strings.TrimSpace(strings.Join(args, " "))
	if len(description) == 0 {
		return errors.New("can not add item with empty description")
	}
	newItem, err := todotxt.BuildItem(
		todotxt.WithDescription(strings.TrimSpace(fmt.Sprintf("%s %s %s", a.def.Prefix, description, a.def.Suffix))),
		todotxt.WithCreationDate(time.Now()),
	)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}
	err = list.Add(newItem)
	if err != nil {
		return fmt.Errorf("could not add task: %w", err)
	}
	list.Reindex()
	fmt.Fprintf(cmd.OutOrStdout(), "Successfully added task with index %d\n", list.IndexOf(newItem))
	return nil
}
