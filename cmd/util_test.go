package cmd_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fabian-G/quest/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var today = time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)

func BuildTestConfig(t *testing.T) *viper.Viper {
	testTodoDir, err := os.MkdirTemp("", "quest-cmd-test-*")
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.RemoveAll(testTodoDir)
	})
	testTodoFile, err := os.CreateTemp(testTodoDir, "*.todo.txt")
	assert.Nil(t, err)
	defer testTodoFile.Close()
	testDoneFile, err := os.CreateTemp(testTodoDir, "*.done.txt")
	assert.Nil(t, err)
	defer testDoneFile.Close()

	cfg := viper.New()
	cfg.Set(config.TodoFileKey, testTodoFile.Name())
	cfg.Set(config.DoneFileKey, testDoneFile.Name())
	cfg.Set(config.NowFuncKey, func() time.Time {
		return today
	})

	cfg.SetDefault(config.ViewsKey, []any{})
	cfg.SetDefault(config.MacrosKey, []any{})
	cfg.SetDefault(config.QuestScoreKey+".urgency-tag", "due")
	cfg.SetDefault(config.QuestScoreKey+".urgency-begin", 90)
	cfg.SetDefault(config.QuestScoreKey+".min-priority", "E")
	cfg.SetDefault(config.UnknownTagsKey, true)
	return cfg
}

func ActivateRecurrence(v *viper.Viper) {
	v.Set("recurrence.rec-tag", "rec")
}

func BuildTestDi(t *testing.T) *config.Di {
	di := config.Di{}
	di.SetConfig(BuildTestConfig(t))
	return &di
}

func ReadLines(t *testing.T, path string) []string {
	lines, err := ReadLinesE(path)
	assert.Nil(t, err)
	return lines
}

func ReadLinesE(path string) ([]string, error) {
	rawData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSuffix(string(rawData), "\n"), "\n")
	return lines, nil
}
