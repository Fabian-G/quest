package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
)

type notesCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
}

func newNotesCommand(def di.ViewDef) *notesCommand {
	cmd := notesCommand{
		viewDef: def,
	}

	return &cmd
}

func (n *notesCommand) command() *cobra.Command {
	var notesCommand = &cobra.Command{
		Use:     "notes [selectors...]",
		Short:   "Opens (or creates) the note attached to the matching task",
		GroupID: "view-cmd",
		PreRunE: cmdutil.Steps(cmdutil.LoadList),
		RunE:    n.notes,
		// No PostRun needed, because we handle saving manually here
	}
	cmdutil.RegisterSelectionFlags(notesCommand, &n.qql, &n.rng, &n.str, nil)

	var cleanCommand = &cobra.Command{
		Use:     "clean",
		Short:   "Removes all notes that are not referenced by a task in todo.txt or done.txt",
		PreRunE: cmdutil.Steps(cmdutil.LoadList, cmdutil.LoadDoneList),
		RunE:    n.clean,
	}
	notesCommand.AddCommand(cleanCommand)
	return notesCommand
}

func (n *notesCommand) notes(cmd *cobra.Command, args []string) (err error) {
	di := cmd.Context().Value(cmdutil.DiKey).(*di.Container)
	repo := di.TodoTxtRepo()
	editor := di.Editor()
	notesRepo := di.NotesRepo()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(n.viewDef.Query, args, n.qql, n.rng, n.str)
	if err != nil {
		return err
	}
	filteredTasks := selector.Filter(list)
	if len(filteredTasks) == 0 {
		fmt.Println("no matches")
		return nil
	}
	selectedTask := filteredTasks[0]
	if len(filteredTasks) > 1 {
		selection, err := selection.New("Select Note to open", filteredTasks).RunPrompt()
		if err != nil {
			return fmt.Errorf("coul not select task: %w", err)
		}
		selectedTask = selection
	}

	note, err := notesRepo.Get(selectedTask)
	if err != nil {
		return fmt.Errorf("could not get note for selected task: %w", err)
	}

	// Save the list before running the editor, because we expect the user to spent a long time in there
	if err = repo.Save(list); err != nil {
		return err
	}
	return editor.Edit(note)
}

func (n *notesCommand) clean(cmd *cobra.Command, args []string) (err error) {
	di := cmd.Context().Value(cmdutil.DiKey).(*di.Container)
	notesRepo := di.NotesRepo()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	doneList := cmd.Context().Value(cmdutil.DoneListKey).(*todotxt.List)

	cont, err := confirmation.New("You are about to delete all notes not referenced from your todo.txt/done.txt. Continue?", confirmation.No).RunPrompt()
	if err != nil {
		return fmt.Errorf("failed to get user confirmation: %w", err)
	}
	if !cont {
		return nil
	}

	if err := notesRepo.Clean(list, doneList); err != nil {
		return fmt.Errorf("failed to clean notes dir: %w", err)
	}

	fmt.Println("Successfully cleaned notes directory")
	return nil
}
