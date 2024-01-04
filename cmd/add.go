package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type addCommand struct {
	def  di.ViewDef
	prio string
}

func newAddCommand(def di.ViewDef) *addCommand {
	cmd := addCommand{
		def: def,
	}

	return &cmd
}

func (a *addCommand) command() *cobra.Command {
	var addCmd = &cobra.Command{
		Use:      AppName + " add [-p prio] description",
		GroupID:  "view-cmd",
		Short:    "Adds a task to the todo list",
		Example:  "quest add do the dishes",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     a.add,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}

	addCmd.Flags().StringVarP(&a.prio, "priority", "p", "none", "TODO")

	return addCmd
}

func (a *addCommand) add(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	description := strings.TrimSpace(strings.Join(args, " "))
	if len(description) == 0 {
		return errors.New("can not add item with empty description")
	}
	prio, err := todotxt.PriorityFromString(a.prio)
	if err != nil {
		return fmt.Errorf("could not parse priority value %s: %w", a.prio, err)
	}
	newItem, err := todotxt.BuildItem(
		todotxt.WithDescription(strings.TrimSpace(fmt.Sprintf("%s %s %s", a.def.AddPrefix, description, a.def.AddSuffix))),
		todotxt.WithCreationDate(time.Now()),
		todotxt.WithPriority(prio),
	)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}
	err = list.Add(newItem)
	if err != nil {
		return fmt.Errorf("could not add task: %w", err)
	}
	fmt.Printf("Added task #%d\n", list.LineOf(newItem))
	return nil
}
