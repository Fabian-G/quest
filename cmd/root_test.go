package cmd_test

import (
	"os"
	"testing"

	"github.com/Fabian-G/quest/cmd"
	"github.com/Fabian-G/quest/di"
	"github.com/stretchr/testify/assert"
)

func Test_RunsWithEmptyConfigurationFile(t *testing.T) {
	tmpConfig, err := os.CreateTemp("", "empty-quest-test-config-*.toml")
	assert.Nil(t, err)
	tmpConfig.Close()

	cmd, ctx := cmd.Root(&di.Container{
		ConfigFile: tmpConfig.Name(),
	})
	cmd.SetArgs([]string{"--json"})
	err = cmd.ExecuteContext(ctx)
	assert.Nil(t, err)
}

// This test only really makes sense in the CI/CD environment, becasue locally there probably is a configuration file present
func Test_RunsWithoutConfigurationFile(t *testing.T) {
	cmd, ctx := cmd.Root(&di.Container{})
	cmd.SetArgs([]string{"--json"})
	err := cmd.ExecuteContext(ctx)
	assert.Nil(t, err)
}
