package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strconv"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type editCommand struct {
	viewDef   config.ViewDef
	qql       []string
	rng       []string
	str       []string
	sortOrder string
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
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     e.edit,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(editCommand, &e.qql, &e.rng, &e.str)
	editCommand.Flags().StringVarP(&e.sortOrder, "sort", "s", e.viewDef.DefaultSortOrder, "TODO")
	return editCommand
}

func (e *editCommand) edit(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(cmdutil.DiKey).(*config.Di)
	repo := di.TodoTxtRepo()
	cfg := di.Config()
	editor := di.Editor()
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
	sortCompiler := qsort.Compiler{
		TagTypes:        di.TagTypes(),
		ScoreCalculator: di.QuestScoreCalculator(),
	}
	sortFunc, err := sortCompiler.CompileSortFunc(e.sortOrder)
	if err != nil {
		return err
	}
	slices.SortStableFunc(selection, sortFunc)

	filePath, err := e.dumpDescriptionsToTempFile(list, selection)
	if err != nil {
		return err
	}
	defer os.Remove(filePath)
	for {
		if err = editor.Edit(filePath); err != nil {
			return err
		}
		additions, changes, removals, err := e.applyChanges(filePath, list, selection)
		if err == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Items Added:   %d\nItems changed: %d\nItems removed: %d\n", additions, changes, removals)
			return nil
		}
		if !cmdutil.AskRetry(cmd, err) {
			return err
		}
		if err := repo.Rollback(list); err != nil {
			return e.handleRollbackFailure(path.Dir(cfg.GetString(config.TodoFileKey)), filePath, err)
		}
		selection = selector.Filter(list)
		slices.SortStableFunc(selection, sortFunc)
	}
}

func (e *editCommand) handleRollbackFailure(backupDir string, changeFilePath string, base error) error {
	fmt.Printf("retry not possible: %s\n", base)
	changeFile, err := os.Open(changeFilePath)
	if err != nil {
		return fmt.Errorf("could not save backup of changes: %w", err)
	}
	defer changeFile.Close()
	changeBackup, err := os.CreateTemp(backupDir, "quest-edit-*.todo.txt")
	if err != nil {
		return fmt.Errorf("could not save backup of changes: %w", err)
	}
	defer changeBackup.Close()
	_, err = io.Copy(changeBackup, changeFile)
	if err != nil {
		return fmt.Errorf("could not save backup of changes: %w", err)
	}
	fmt.Printf("a copy of your changes has been saved at %s, but you have to apply them manually", changeBackup.Name())
	return err
}

func (e *editCommand) dumpDescriptionsToTempFile(list *todotxt.List, items []*todotxt.Item) (string, error) {
	if err := setObjectIdTag(list, items); err != nil {
		return "", err
	}
	tmpFile, err := os.CreateTemp("", "quest-edit-*.todo.txt")
	if err != nil {
		return "", fmt.Errorf("could not create tmp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
	}()
	writer := bufio.NewWriter(tmpFile)
	if err = todotxt.DefaultEncoder.Encode(writer, items); err != nil {
		return "", err
	}

	if err := clearObjectIdTag(list, items); err != nil {
		return "", err
	}

	return tmpFile.Name(), writer.Flush()
}

func (e *editCommand) applyChanges(tmpFile string, list *todotxt.List, selection []*todotxt.Item) (added int, changed int, removed int, err error) {
	file, err := os.Open(tmpFile)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()
	changeList, err := todotxt.DefaultDecoder.Decode(file)
	if err != nil {
		return 0, 0, 0, err
	}

	changedItems, addedItems, err := mapToIds(changeList, len(selection))
	if err != nil {
		return 0, 0, 0, err
	}

	for _, item := range addedItems {
		err := list.Add(item)
		if err != nil {
			return 0, 0, 0, err
		}
		added++
	}

	deletedItems := make([]*todotxt.Item, len(selection)) // what remains in this list after the loop will be deleted
	copy(deletedItems, selection)
	for _, item := range changedItems {
		deletedItems[item.id] = nil
		if item.item.Equals(selection[item.id]) {
			continue
		}
		changed++
		if err := selection[item.id].Apply(item.item); err != nil {
			return 0, 0, 0, err
		}
	}

	for _, i := range deletedItems {
		if i == nil {
			continue
		}
		removed++
		if err := list.Remove(list.IndexOf(i)); err != nil {
			return 0, 0, 0, err
		}
	}

	return
}

func getEditId(item *todotxt.Item) (int, error) {
	tagValues := item.Tags()[config.InternalEditTag]
	if len(tagValues) == 0 {
		return -1, nil
	}
	id, err := strconv.Atoi(tagValues[0])
	if err != nil {
		return 0, fmt.Errorf("encountered invalid %s tag. Dot not change them: %w", config.InternalEditTag, err)
	}
	return id, nil
}

func setObjectIdTag(list *todotxt.List, items []*todotxt.Item) error {
	return list.Secret(func() error {
		for i, item := range items {
			if err := item.SetTag(config.InternalEditTag, strconv.Itoa(i)); err != nil {
				return err
			}
		}
		return nil
	})
}

func clearObjectIdTag(list *todotxt.List, items []*todotxt.Item) error {
	return list.Secret(func() error {
		for _, item := range items {
			if err := item.SetTag(config.InternalEditTag, ""); err != nil {
				return err
			}
		}
		return nil
	})
}

type itemWithId struct {
	id   int
	item *todotxt.Item
}

func mapToIds(items []*todotxt.Item, maxId int) ([]itemWithId, []*todotxt.Item, error) {
	idMap := make([]itemWithId, 0) // slice instead of map, because we must retain the order
	noId := make([]*todotxt.Item, 0)
	for _, item := range items {
		id, err := getEditId(item)
		if err != nil {
			return nil, nil, err
		}
		if id == -1 {
			noId = append(noId, item)
		} else if slices.ContainsFunc(idMap, func(iwi itemWithId) bool { return iwi.id == id }) {
			return nil, nil, fmt.Errorf("encountered duplicate id %d. Do not change the %s tag", id, config.InternalEditTag)
		} else {
			idMap = append(idMap, itemWithId{id: id, item: item})

			item.SetTag(config.InternalEditTag, "")
		}
	}
	return idMap, noId, nil
}
