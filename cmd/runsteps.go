package cmd

import (
	"context"
	"fmt"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
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
