package cmd_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fabian-G/quest/di"
	"github.com/stretchr/testify/assert"
)

var today = time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)

func BuildTestConfig(t *testing.T, opts ...func(di.Config) di.Config) di.Config {
	testTodoDir, err := os.MkdirTemp("", "quest-cmd-test-*")
	assert.Nil(t, err)
	t.Cleanup(func() {
		os.RemoveAll(testTodoDir)
	})
	testTodoFile, err := os.CreateTemp(testTodoDir, "*.todo.txt")
	assert.Nil(t, err)
	assert.Nil(t, testTodoFile.Close())
	testDoneFile, err := os.CreateTemp(testTodoDir, "*.done.txt")
	assert.Nil(t, err)
	assert.Nil(t, testDoneFile.Close())

	cfg := di.Config{}
	cfg.TodoFile = testTodoFile.Name()
	cfg.DoneFile = testDoneFile.Name()
	cfg.NowFunc = func() time.Time {
		return today
	}

	cfg.QuestScore.MinPriority = "E"
	cfg.QuestScore.UrgencyBegin = 90
	cfg.QuestScore.UrgencyTags = []string{"due"}
	cfg.UnknownTags = true
	cfg.Tags = map[string]di.TagDef{
		di.InternalEditTag: {
			Type:     "int",
			Humanize: false,
		},
	}

	for _, o := range opts {
		cfg = o(cfg)
	}
	return cfg
}

func WithoutUnknownTags(c di.Config) di.Config {
	c.UnknownTags = false
	return c
}

func WithRecurrence(cfg di.Config) di.Config {
	cfg.Recurrence.RecTag = "rec"
	cfg.Recurrence.DueTag = "due"
	cfg.Recurrence.ThresholdTag = "t"
	return cfg
}

func WithTag(key string, typ string) func(di.Config) di.Config {
	return func(c di.Config) di.Config {
		c.Tags[key] = di.TagDef{
			Type:     typ,
			Humanize: false,
		}
		return c
	}
}

func BuildTestDi(t *testing.T, config di.Config) *di.Container {
	di := di.Container{}
	di.SetConfig(config)
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
