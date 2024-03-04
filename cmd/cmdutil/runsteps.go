package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type CtxKey string

var (
	DiKey       CtxKey = "DI"
	ListKey     CtxKey = "list"
	DoneListKey CtxKey = "done-list"
)

func Steps(steps ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range steps {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func ConfigOverrides(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(DiKey).(*di.Container)
	cfg := di.Config()

	file := cmd.Root().PersistentFlags().Lookup("file")
	if file.Changed {
		cfg.TodoFile = file.Value.String()
	}

	di.SetConfig(cfg)
	return nil
}

func LoadList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*di.Container).TodoTxtRepo()
	list, err := repo.Read()
	if err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), ListKey, list))
	return nil
}

func SaveList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*di.Container).TodoTxtRepo()
	list := cmd.Context().Value(ListKey).(*todotxt.List)
	if err := repo.Save(list); err != nil {
		return fmt.Errorf("could not save todo file: %w", err)
	}
	return repo.Close()
}

func LoadDoneList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*di.Container).DoneTxtRepo()
	list, err := repo.Read()
	if err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), DoneListKey, list))
	return nil
}

func SaveDoneList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*di.Container).DoneTxtRepo()
	list := cmd.Context().Value(DoneListKey).(*todotxt.List)
	if err := repo.Save(list); err != nil {
		return fmt.Errorf("could not save done file: %w", err)
	}
	return repo.Close()
}

func createFileIfNotExists(file string) error {
	stat, err := os.Stat(file)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.MkdirAll(path.Dir(file), 0777); err != nil {
			return err
		}
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		return f.Close()
	case err != nil:
		return err
	case !stat.Mode().IsRegular():
		return fmt.Errorf("provided file %s is not a regular file", file)
	}
	return nil
}

func EnsureTodoFileExits(cmd *cobra.Command, args []string) error {
	v := cmd.Context().Value(DiKey).(*di.Container).Config()
	file := v.TodoFile
	return createFileIfNotExists(file)
}

func EnsureDoneFileExists(cmd *cobra.Command, args []string) error {
	v := cmd.Context().Value(DiKey).(*di.Container).Config()
	file := v.DoneFile
	return createFileIfNotExists(file)
}

func EnsureNotesDirExists(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(DiKey).(*di.Container)
	if di.NotesRepo() == nil {
		return nil // Feature not activated (so don't bother)
	}
	err := os.Mkdir(di.Config().Notes.Dir, 0777)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	return err
}

func RegisterMacros(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(DiKey).(*di.Container)
	for _, macro := range di.Config().Macros {
		err := qselect.RegisterMacro(macro.Name, macro.Query, macro.InDTypes(), qselect.DType(macro.ResultType), macro.InjectIt)
		if err != nil {
			return fmt.Errorf("could not register macro %s: %w", macro.Name, err)
		}
	}
	return nil
}

func SyncConflictProtection(cmd *cobra.Command, args []string) error {
	v := cmd.Context().Value(DiKey).(*di.Container).Config()
	file := v.TodoFile
	filesInDir, err := os.ReadDir(path.Dir(file))
	if err != nil {
		return fmt.Errorf("could not check for sync conflicts: %w", err)
	}

	base := path.Base(file)
	extension := path.Ext(file)
	name := strings.TrimSuffix(base, extension)
	syncthingConflictMatcher := regexp.MustCompile(fmt.Sprintf("^%s\\.sync-conflict-.*%s$", name, extension))
	questOLockConflictMatcher := regexp.MustCompile(fmt.Sprintf("^%s\\.quest-conflict-.*%s$", name, extension))
	conflicts := make([]string, 0)
	for _, f := range filesInDir {
		if syncthingConflictMatcher.MatchString(path.Base(f.Name())) || questOLockConflictMatcher.MatchString(path.Base(f.Name())) {
			conflicts = append(conflicts, path.Join(path.Dir(file), f.Name()))
		}
	}
	if len(conflicts) == 0 {
		return nil
	}

	fmt.Printf("The following sync conflicts have been detected:\n")
	for _, c := range conflicts {
		fmt.Printf("- %s\n", c)
	}
	fmt.Println("\nPlease merge manually and then remove the specified files.")
	return errors.New("sync conflict")
}
