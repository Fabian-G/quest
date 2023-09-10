/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type ctxKey string

var (
	queryKey     ctxKey = "query"
	diKey        ctxKey = "DI"
	listKey      ctxKey = "list"
	selectionKey ctxKey = "selection"
)

var (
	folSearch    []string
	rngSearch    []string
	stringSearch []string
)

var rootCmd = &cobra.Command{
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initSteps := []func(cmd *cobra.Command, args []string) error{
			parseTaskSelection,
			initDI,
			loadList,
			selectItems,
		}
		for _, f := range initSteps {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		afterSteps := []func(cmd *cobra.Command, args []string) error{
			saveList,
		}
		for _, f := range afterSteps {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func parseTaskSelection(cmd *cobra.Command, args []string) error {
	selectors := make([]query.Func, 0)
	for _, arg := range args {
		q, err := query.Compile(arg, query.Guess)
		if err != nil {
			return fmt.Errorf("could not compile query %s. Try using -q,-r or -s explicitly instead of positional args: %w", arg, err)
		}
		selectors = append(selectors, q)
	}
	for _, f := range folSearch {
		q, err := query.Compile(f, query.FOL)
		if err != nil {
			return fmt.Errorf("could not compile FOL query %s: %w", f, err)
		}
		selectors = append(selectors, q)
	}
	for _, r := range rngSearch {
		q, err := query.Compile(r, query.Range)
		if err != nil {
			return fmt.Errorf("could not compile range query %s: %w", r, err)
		}
		selectors = append(selectors, q)
	}
	for _, s := range stringSearch {
		q, err := query.Compile(s, query.StringSearch)
		if err != nil {
			return fmt.Errorf("could not compile string search query %s: %w", s, err)
		}
		selectors = append(selectors, q)
	}

	cmd.SetContext(context.WithValue(cmd.Context(), queryKey, query.And(selectors...)))

	return nil
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

func selectItems(cmd *cobra.Command, args []string) error {
	query := cmd.Context().Value(queryKey).(query.Func)
	list := cmd.Context().Value(listKey).(*todotxt.List)
	selection := query.Filter(list)
	cmd.SetContext(context.WithValue(cmd.Context(), selectionKey, selection))
	return nil
}

func saveList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(diKey).(*config.Di).TodoTxtRepo()
	list := cmd.Context().Value(listKey).(*todotxt.List)
	return repo.Save(list)
}

func init() {
	listCmd.Flags().StringArrayVarP(&folSearch, "fol", "q", nil, "TODO")
	listCmd.Flags().StringArrayVarP(&rngSearch, "range", "r", nil, "TODO")
	listCmd.Flags().StringArrayVarP(&stringSearch, "string", "s", nil, "TODO")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})
}
