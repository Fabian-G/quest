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

	di.SetEditor(editSteps(complete(1), shuffle()))

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

	di.SetEditor(editAttempts(appendInvalidLine(), editSteps(removeLine(3), complete(1), shuffle())))

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

func Test_TagExpansionIsEvaluatedInOrderOfTheEditFileEvenAfterRollback(t *testing.T) {
	di := BuildTestDi(t)
	di.Config().Set(config.TagsKey, map[string]string{
		"order": "int",
	})
	todoFile := di.Config().GetString(config.TodoFileKey)
	todos := `+P1 T1
+P1 T2
+P1 T3`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editAttempts(
		editSteps(appendOnEach("order:pmax+10"), appendInvalidLine()),
		editSteps(removeLine(3), reverse()),
	))

	err := cmd.Execute(di, []string{"edit", "-s", ""})

	assert.Nil(t, err)
	lines := ReadLines(t, todoFile)
	assert.Len(t, lines, 3)
	assert.ElementsMatch(t, ReadLines(t, todoFile), []string{
		"+P1 T1 order:30",
		"+P1 T2 order:20",
		"+P1 T3 order:10",
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
func appendOnEach(suffix string) config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		for i, l := range lines {
			lines[i] = fmt.Sprintf("%s %s", l, suffix)
		}
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
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

func complete(completeTask int) config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines[completeTask] = fmt.Sprintf("x %s", lines[completeTask])
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func reverse() config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		slices.Reverse(lines)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func shuffle() config.Editor {
	return config.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		dest := make([]string, len(lines))
		perm := rand.Perm(len(lines))
		for i, v := range perm {
			dest[v] = lines[i]
		}
		return os.WriteFile(path, []byte(strings.Join(dest, "\n")), 0644)
	})
}
