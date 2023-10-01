package cmd

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type editCommand struct {
	viewDef   di.ViewDef
	qql       []string
	rng       []string
	str       []string
	sortOrder []string
}

func newEditCommand(def di.ViewDef) *editCommand {
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
	editCommand.Flags().StringSliceVarP(&e.sortOrder, "sort", "s", e.viewDef.Sort, "TODO")
	return editCommand
}

func (e *editCommand) edit(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(cmdutil.DiKey).(*di.Container)
	editor := di.Editor()
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(e.viewDef.Query, args, e.qql, e.rng, e.str)
	if err != nil {
		return err
	}
	selection := selector.Filter(list)
	if len(selection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}
	sortCompiler := di.SortCompiler()
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
	list.Snapshot()
	for {
		if err = editor.Edit(filePath); err != nil {
			return err
		}
		additions, changes, removals, err := e.applyChanges(filePath, list, selection, di.Config().NowFunc)
		if err == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Items Added:   %d\nItems changed: %d\nItems removed: %d\n", additions, changes, removals)
			return nil
		}
		if !cmdutil.AskRetry(cmd, err) {
			return err
		}
		list.Reset()
	}
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

func (e *editCommand) applyChanges(tmpFile string, list *todotxt.List, selection []*todotxt.Item, now func() time.Time) (added int, changed int, removed int, err error) {
	file, err := os.Open(tmpFile)
	if err != nil {
		return 0, 0, 0, err
	}
	defer file.Close()
	changeList, err := todotxt.DefaultDecoder.Decode(file)
	if err != nil {
		return 0, 0, 0, err
	}

	changedItems, err := mapToIds(changeList, len(selection))
	if err != nil {
		return 0, 0, 0, err
	}

	deletedItems := make([]*todotxt.Item, len(selection)) // what remains in this list after the loop will be deleted
	copy(deletedItems, selection)
	for _, item := range changedItems {
		if item.id == -1 {
			if item.item.CreationDate() == nil {
				if err := item.item.SetCreation(now()); err != nil {
					return 0, 0, 0, err
				}
			}
			err := list.Add(item.item)
			if err != nil {
				return 0, 0, 0, err
			}
			added++
			continue
		}
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
		if err := list.Remove(list.LineOf(i)); err != nil {
			return 0, 0, 0, err
		}
	}

	return
}

func getEditId(item *todotxt.Item) (int, error) {
	tagValues := item.Tags()[di.InternalEditTag]
	if len(tagValues) == 0 {
		return -1, nil
	}
	id, err := strconv.Atoi(tagValues[0])
	if err != nil {
		return 0, fmt.Errorf("encountered invalid %s tag. Dot not change them: %w", di.InternalEditTag, err)
	}
	return id, nil
}

func setObjectIdTag(list *todotxt.List, items []*todotxt.Item) error {
	return list.Secret(func() error {
		for i, item := range items {
			if err := item.SetTag(di.InternalEditTag, strconv.Itoa(i)); err != nil {
				return err
			}
		}
		return nil
	})
}

func clearObjectIdTag(list *todotxt.List, items []*todotxt.Item) error {
	return list.Secret(func() error {
		for _, item := range items {
			if err := item.SetTag(di.InternalEditTag, ""); err != nil {
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

func mapToIds(items []*todotxt.Item, maxId int) ([]itemWithId, error) {
	idMap := make([]itemWithId, 0) // slice instead of map, because we must retain the order
	for _, item := range items {
		id, err := getEditId(item)
		if err != nil {
			return nil, err
		}
		if id == -1 {
			idMap = append(idMap, itemWithId{id: -1, item: item})
		} else if slices.ContainsFunc(idMap, func(iwi itemWithId) bool { return iwi.id == id }) {
			return nil, fmt.Errorf("encountered duplicate id %d. Do not change the %s tag", id, di.InternalEditTag)
		} else {
			idMap = append(idMap, itemWithId{id: id, item: item})
			item.SetTag(di.InternalEditTag, "")
		}
	}
	return idMap, nil
}
