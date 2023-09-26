package cmd_test

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Fabian-G/quest/cmd"
	"github.com/Fabian-G/quest/config"
	"github.com/stretchr/testify/assert"
)

func Test_EditsAreCorrectEvenWhenShufflingLines(t *testing.T) {
	di := BuildTestDi(t)
	ActivateRecurrence(di.Config())
	todoFile := di.Config().GetString(config.TodoFileKey)
	todos := `x a done task
A recurring task rec:+1w due:2020-01-01
Another task`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(completeAndShuffle(1))

	err := cmd.Execute(di, []string{"edit", "-s", ""})

	assert.Nil(t, err)
	lines := ReadLines(t, todoFile)
	assert.Len(t, lines, 4)
	assert.ElementsMatch(t, ReadLines(t, todoFile), []string{
		"x a done task",
		"x A recurring task rec:+1w due:2020-01-01",
		"2022-02-02 A recurring task rec:+1w due:2020-01-08",
		"Another task",
	})
}

func Test_RecurrenceIsTriggeredAfterRollback(t *testing.T) {
	di := BuildTestDi(t)
	ActivateRecurrence(di.Config())
	todoFile := di.Config().GetString(config.TodoFileKey)
	todos := `x a done task
A recurring task rec:+1w due:2020-01-01
Another task`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editAttempts(appendInvalidLine(), editSteps(removeLine(3), completeAndShuffle(1))))

	err := cmd.Execute(di, []string{"edit", "-s", ""})

	assert.Nil(t, err)
	lines := ReadLines(t, todoFile)
	assert.Len(t, lines, 4)
	assert.ElementsMatch(t, ReadLines(t, todoFile), []string{
		"x a done task",
		"x A recurring task rec:+1w due:2020-01-01",
		"2022-02-02 A recurring task rec:+1w due:2020-01-08",
		"Another task",
	})
}

func editAttempts(editors ...config.Editor) config.Editor {
	invocationCount := 0
	return config.EditorFunc(func(path string) error {
		err := editors[invocationCount].Edit(path)
		invocationCount++
		return err
	})
}

func editSteps(editors ...config.Editor) config.Editor {
	return config.EditorFunc(func(path string) error {
		for _, e := range editors {
			err := e.Edit(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func appendInvalidLine() config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines = append(lines, "2022-13-01 Hello World")
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func removeLine(idx int) config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines = slices.Delete(lines, idx, idx+1)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func completeAndShuffle(completeTask int) config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines[completeTask] = fmt.Sprintf("x %s", lines[completeTask])
		dest := make([]string, len(lines))
		perm := rand.Perm(len(lines))
		for i, v := range perm {
			dest[v] = lines[i]
		}
		return os.WriteFile(path, []byte(strings.Join(dest, "\n")), 0644)
	})
}
