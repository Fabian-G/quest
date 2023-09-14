package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ctxKey string

var (
	queryKey ctxKey = "query"
	diKey    ctxKey = "DI"
	listKey  ctxKey = "list"
)

var (
	qqlSearch    []string
	rngSearch    []string
	stringSearch []string
)

var rootCmd = &cobra.Command{
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initSteps := []func(cmd *cobra.Command, args []string) error{
			parseTaskSelection,
			initDI,
			loadList,
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
	for _, f := range qqlSearch {
		q, err := query.Compile(f, query.QQL)
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

func saveList(cmd *cobra.Command, args []string) error {
	repo := cmd.Context().Value(diKey).(*config.Di).TodoTxtRepo()
	list := cmd.Context().Value(listKey).(*todotxt.List)
	return repo.Save(list)
}

func init() {
	rootCmd.PersistentFlags().StringArrayVarP(&qqlSearch, "qql", "q", nil, "TODO")
	rootCmd.PersistentFlags().StringArrayVarP(&rngSearch, "range", "r", nil, "TODO")
	rootCmd.PersistentFlags().StringArrayVarP(&stringSearch, "word", "w", nil, "TODO")
	rootCmd.AddGroup(&cobra.Group{
		ID:    "query",
		Title: "Query",
	})

	rootCmd.AddCommand(listCmd(viewDef{
		name:              "list",
		defaultSelection:  "",
		defaultProjection: view.StarProjection,
		defaultSortOrder:  "+done,+creation,+description",
		defaultOutputMode: interactiveOutput,
		defaultClean:      nil,
	}))
	views := viper.GetStringMap("view")
	for name, viewDefA := range views {
		viewDefM, ok := viewDefA.(map[string]any)
		if !ok {
			log.Fatalf("error in config file. expected view definition in section [view.%s], but got %T", name, viewDefA)
		}
		var (
			selection  string   = ""
			projection string   = view.StarProjection
			sortOrder  string   = "+done,+cretion,+description"
			output     string   = interactiveOutput
			clean      []string = nil
		)
		if s, ok := viewDefM["query"]; ok {
			selection = s.(string)
		}
		if p, ok := viewDefM["projection"]; ok {
			projection = p.(string)
		}
		if s, ok := viewDefM["sort"]; ok {
			sortOrder = s.(string)
		}
		if i, ok := viewDefM["output"]; ok {
			output = i.(string)
		}
		if c, ok := viewDefM["clean"]; ok {
			cleanS := c.(string)
			clean = strings.Split(cleanS, ",")
		}
		rootCmd.AddCommand(listCmd(viewDef{
			name:              name,
			defaultSelection:  selection,
			defaultProjection: projection,
			defaultSortOrder:  sortOrder,
			defaultOutputMode: output,
			defaultClean:      clean,
		}))
	}
}
