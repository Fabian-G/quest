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
	defer os.Remove(filePath)
	for {
		if err = e.startEditor(cfg.GetString(config.Editor), filePath); err != nil {
			return err
		}
		changes, removals, err := e.applyChanges(filePath, writtenLines, list, selection)
		if err == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Items edited:  %d\nItems removed: %d\n", changes, removals)
			return nil
		}
		if !askRetry(cmd, err) {
			return err
		}
	}
}

func askRetry(cmd *cobra.Command, err error) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "your changes are invalid: %s\n", err)
	fmt.Fprint(cmd.OutOrStdout(), "Retry? (Y/n) ")
	var answer string
	fmt.Fscanln(cmd.InOrStdin(), &answer)
	return strings.ToLower(answer) != "n"
}

func (e *editCommand) dumpDescriptionsToTempFile(items []*todotxt.Item) (string, int, error) {
	tmpFile, err := os.CreateTemp("", "quest-edit-*.todo.txt")
	if err != nil {
		return "", 0, fmt.Errorf("could not create tmp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
	}()
	writer := bufio.NewWriter(tmpFile)
	if err = todotxt.DefaultEncoder.Encode(writer, items); err != nil {
		return "", 0, err
	}

	return tmpFile.Name(), len(items), writer.Flush()
}

func (e *editCommand) startEditor(editorCmd string, path string) error {
	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor command failed: %w", err)
	}
	return nil
}

func (e *editCommand) applyChanges(tmpFile string, expectedLines int, list *todotxt.List, selection []*todotxt.Item) (int, int, error) {
	file, err := os.Open(tmpFile)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		file.Close()
	}()
	changeList, err := todotxt.DefaultDecoder.Decode(file)
	if err != nil {
		return 0, 0, err
	}
	if len(changeList) != expectedLines {
		return 0, 0, fmt.Errorf("expected %d lines, but got %d. Do not delete, add or reorder lines when editing", expectedLines, len(changeList))
	}

	var removedItems, changedItems int
	for idx, item := range changeList {
		if strings.TrimSpace(item.Description()) == "" {
			removedItems++
			list.Remove(list.IndexOf(selection[idx]))
		}
		before := *selection[idx]
		if err := selection[idx].Apply(item); err != nil {
			return 0, 0, err
		}
		if !before.Equals(selection[idx]) {
			changedItems++
		}
	}

	return changedItems, removedItems, nil
}
