package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ctxKey string

var (
	diKey   ctxKey = "DI"
	listKey ctxKey = "list"
)

func steps(steps ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range steps {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func initDI(cmd *cobra.Command, args []string) error {
	di := &config.Di{}
	cmd.SetContext(context.WithValue(cmd.Context(), diKey, di))
	return nil
}

func loadList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(diKey).(*config.Di).TodoTxtRepo()
	list, err := repo.Read()
	if err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(cmd.Context(), listKey, list))
	return nil
}

func saveList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(diKey).(*config.Di).TodoTxtRepo()
	list := cmd.Context().Value(listKey).(*todotxt.List)
	if err := repo.Save(list); err != nil {
		return fmt.Errorf("could not save todo file: %w", err)
	}
	return nil
}

func ensureTodoFileExits(cmd *cobra.Command, args []string) error {
	file := viper.GetString(config.TodoFile)
	stat, err := os.Stat(file)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.MkdirAll(path.Dir(file), 0777); err != nil {
			return err
		}
		if _, err := os.Create(file); err != nil {
			return err
		}
	case err != nil:
		return err
	case !stat.Mode().IsRegular():
		return fmt.Errorf("provided file %s is not a regular file", file)
	}
	return nil
}
