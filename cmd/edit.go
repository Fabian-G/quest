package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type editCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
}

func newEditCommand(def config.ViewDef) *editCommand {
	cmd := editCommand{
		viewDef: def,
	}

	return &cmd
}

func (e *editCommand) command() *cobra.Command {
	var editCommand = &cobra.Command{
		Use:      "edit",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     e.edit,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(editCommand, &e.qql, &e.rng, &e.str)
	return editCommand
}

func (e *editCommand) edit(cmd *cobra.Command, args []string) error {
	cfg := cmd.Context().Value(cmdutil.DiKey).(*config.Di).Config()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(e.viewDef.DefaultQuery, args, e.qql, e.rng, e.str)
	if err != nil {
		return err
	}
	selection := selector.Filter(list)
	if len(selection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}

	filePath, writtenLines, err := e.dumpDescriptionsToTempFile(selection)
	if err != nil {
		return err
	}
	defer func() {
		os.Remove(filePath)
	}()
	if err = e.startEditor(cfg.GetString(config.Editor), filePath); err != nil {
		return err
	}
	changes, removals, err := e.applyChanges(filePath, writtenLines, list, selection)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Items edited:  %d\nItems removed: %d\n", changes, removals)
	return nil
}

func (e *editCommand) dumpDescriptionsToTempFile(items []*todotxt.Item) (string, int, error) {
	tmpFile, err := os.CreateTemp("", "quest-edit-*.txt")
	if err != nil {
		return "", 0, fmt.Errorf("could not create tmp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
	}()
	var writtenLines int
	writer := bufio.NewWriter(tmpFile)
	for _, i := range items {
		writer.WriteString(i.Description())
		writer.WriteString("\n")
		writtenLines++
	}

	return tmpFile.Name(), writtenLines, writer.Flush()
}

func (e *editCommand) startEditor(editorCmd string, path string) error {
	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (e *editCommand) applyChanges(tmpFile string, expectedLines int, list *todotxt.List, selection []*todotxt.Item) (int, int, error) {
	file, err := os.Open(tmpFile)
	if err != nil {
		return 0, 0, err
	}
	scanner := bufio.NewScanner(file)
	var removedItems, changedLines, totalLines int
	for scanner.Scan() {
		totalLines++
		if totalLines > len(selection) {
			continue
		}
		currentItem := selection[totalLines-1]
		line := scanner.Text()
		if line == currentItem.Description() {
			continue
		}
		if strings.TrimSpace(line) == "" {
			removedItems++
			err := list.Remove(list.IndexOf(currentItem))
			if err != nil {
				return 0, 0, err
			}
			continue
		}
		err := currentItem.EditDescription(line)
		if err != nil {
			return 0, 0, err
		}
		changedLines++
	}

	if totalLines != expectedLines {
		return 0, 0, fmt.Errorf("expected %d lines, but got %d. Do not delete, add or reorder lines when editing", expectedLines, totalLines)
	}
	return changedLines, removedItems, nil
}
