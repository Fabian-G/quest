package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Fabian-G/quest/di"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
)

var (
	//go:embed presets/todo.txt.tmpl
	todoTxtPreset string
	//go:embed presets/ext-todo.txt.tmpl
	extTodoTxtPreset string
	//go:embed presets/gtd.tmpl
	gtdPreset string
	//go:embed presets/devs-choice.tmpl
	devsChoicePreset string
)

type preset struct {
	name        string
	description string
	template    string
}

func (p preset) String() string {
	return fmt.Sprintf("%s%s\n", p.name, p.description)
}

var presets []preset = []preset{
	{
		name: "todo.txt",
		description: `
	Simple todo.txt preset for simplest of needs.`,
		template: todoTxtPreset,
	},
	{
		name: "extended todo.txt",
		description: `
	Todo.txt with extensions like due/threshold dates and recurrence.
	This is the config for your if you want to start simple and build on top of that.`,
		template: extTodoTxtPreset,
	},
	{
		name: "GTD",
		description: `
	One method to implement a GTD workflow with quest.`,
		template: gtdPreset,
	},
	{
		name: "Dev's Choice",
		description: `
	The config file used by the dev.`,
		template: devsChoicePreset,
	},
}

type initCommand struct {
}

func newInitCommand() *initCommand {
	cmd := initCommand{}

	return &cmd
}

func (i *initCommand) command() *cobra.Command {
	var initCommand = &cobra.Command{
		Use:     "init",
		Short:   "Helps you to get started by initializing your config",
		GroupID: "global-cmd",
		RunE:    i.init,
	}
	return initCommand
}

func (i *initCommand) init(cmd *cobra.Command, args []string) error {
	configPath := di.DefaultConfigLocation()

	_, err := os.Stat(configPath)
	if err == nil {
		return fmt.Errorf("the config file at location %s already exists. Remove it to use init again.", configPath)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not run stat: %w", err)
	}

	prompt := selection.New("Choose your preset:", presets)
	prompt.SelectedChoiceStyle = nil
	prompt.Filter = nil
	prompt.FinalChoiceStyle = func(s *selection.Choice[preset]) string {
		return selection.DefaultFinalChoiceStyle(&selection.Choice[string]{
			String: s.Value.name,
			Value:  s.Value.name,
		})
	}
	prompt.KeyMap.Down = append(prompt.KeyMap.Down, "j")
	prompt.KeyMap.Up = append(prompt.KeyMap.Up, "k")
	preset, err := prompt.RunPrompt()
	if err != nil {
		return fmt.Errorf("preset selection failed: %w", err)
	}

	type choiceData struct {
		// Other values to be put in the template (empty for now)
	}
	var choices choiceData

	templ, err := template.New("config").Parse(preset.template)
	if err != nil {
		return fmt.Errorf("could not parse embedded template: %w", err)
	}

	info, err := os.Stat(path.Dir(configPath))
	if err == nil && !info.Mode().IsDir() {
		return fmt.Errorf("config home exists, but is not a directory: %w", err)
	} else if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path.Dir(configPath), 0777); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	out, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer out.Close()
	if err = templ.Execute(out, choices); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}
