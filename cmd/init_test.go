package cmd

import (
	"os"
	"testing"
	"text/template"

	"github.com/Fabian-G/quest/di"
	"github.com/stretchr/testify/assert"
)

func Test_PresetsCanBeParsed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}
	viewsToTest := [][]string{
		{"", "all"},
		{"", "all"},
		{"", "inbox", "next", "waiting", "someday", "scheduled"},
		{"", "inbox", "next", "waiting", "someday", "scheduled", "today", "top"},
	}
	for i, preset := range presets {
		i := i
		t.Run(preset.name, func(t *testing.T) {
			template, err := template.New("test-config").Parse(preset.template)
			assert.NoError(t, err)
			tmpConfig, err := os.CreateTemp("", "quest-preset-config-*.toml")
			t.Cleanup(func() {
				tmpConfig.Close()
				os.Remove(tmpConfig.Name())
			})
			assert.Nil(t, err)
			err = template.Execute(tmpConfig, struct{}{})
			assert.NoError(t, err)
			tmpConfig.Close()
			for _, view := range viewsToTest[i] {
				cmd, ctx := Root(&di.Container{
					ConfigFile: tmpConfig.Name(),
				})
				cmd.SetArgs([]string{view, "--json"})
				err = cmd.ExecuteContext(ctx)
				assert.Nil(t, err)
			}
		})
	}
}
