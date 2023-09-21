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

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type CtxKey string

var (
	DiKey   CtxKey = "DI"
	ListKey CtxKey = "list"
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

func LoadList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*config.Di).TodoTxtRepo()
	list, err := repo.Read()
	if err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), ListKey, list))
	return nil
}

func SaveList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(DiKey).(*config.Di).TodoTxtRepo()
	list := cmd.Context().Value(ListKey).(*todotxt.List)
	if err := repo.Save(list); err != nil {
		return fmt.Errorf("could not save todo file: %w", err)
	}
	return nil
}

func EnsureTodoFileExits(cmd *cobra.Command, args []string) error {
	v := cmd.Context().Value(DiKey).(*config.Di).Config()
	file := v.GetString(config.TodoFile)
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
		f.Close()
	case err != nil:
		return err
	case !stat.Mode().IsRegular():
		return fmt.Errorf("provided file %s is not a regular file", file)
	}
	return nil
}

func RegisterMacros(cmd *cobra.Command, args []string) error {
	di := cmd.Context().Value(DiKey).(*config.Di)
	for _, macro := range di.MacroDefs() {
		err := qselect.RegisterMacro(macro.Name, macro.Query, macro.InTypes, macro.ResultType, macro.InjectIt)
		if err != nil {
			return err
		}
	}
	return nil
}

func SyncConflictProtection(cmd *cobra.Command, args []string) error {
	v := cmd.Context().Value(DiKey).(*config.Di).Config()
	file := v.GetString(config.TodoFile)
	filesInDir, err := os.ReadDir(path.Dir(file))
	if err != nil {
		return fmt.Errorf("could check for sync conflicts: %w", err)
	}

	base := path.Base(file)
	extension := path.Ext(file)
	name := strings.TrimSuffix(base, extension)
	conflictMatcher := regexp.MustCompile(fmt.Sprintf("^%s\\.sync-conflict-.*%s$", name, extension))
	conflicts := make([]string, 0)
	for _, f := range filesInDir {
		if conflictMatcher.MatchString(path.Base(f.Name())) {
			conflicts = append(conflicts, path.Join(path.Dir(file), f.Name()))
		}
	}
	if len(conflicts) == 0 {
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "The following sync conflicts have been detected:\n")
	for _, c := range conflicts {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", c)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nContinue anyway? (y/N) ")
	var response string
	fmt.Fscanln(cmd.InOrStdin(), &response)
	if strings.ToUpper(strings.TrimSpace(response)) == "Y" {
		return nil
	}
	return fmt.Errorf("cancelled by user")
}
