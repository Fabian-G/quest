package cmd_test

import (
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Fabian-G/quest/cmd"
	"github.com/Fabian-G/quest/di"
	"github.com/stretchr/testify/assert"
)

func Test_EditsAreCorrectEvenWhenShufflingLines(t *testing.T) {
	di := BuildTestDi(t, BuildTestConfig(t, WithRecurrence))
	todoFile := di.Config().TodoFile
	todos := `x a done task
A recurring task rec:+1w due:2020-01-01
Another task`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editSteps(complete(1), shuffle()))

	cmd, ctx := cmd.Root(di)
	cmd.SetArgs([]string{"edit", "-s", ""})
	err := cmd.ExecuteContext(ctx)

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
	di := BuildTestDi(t, BuildTestConfig(t, WithRecurrence))
	todoFile := di.Config().TodoFile
	todos := `x a done task
A recurring task rec:+1w due:2020-01-01
Another task`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editAttempts(appendInvalidLine(), editSteps(removeLine(3), complete(1), shuffle())))

	cmd, ctx := cmd.Root(di)
	cmd.SetArgs([]string{"edit", "-s", ""})
	cmd.SetIn(strings.NewReader("Y\n"))
	err := cmd.ExecuteContext(ctx)

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
	di := BuildTestDi(t, BuildTestConfig(t, WithoutUnknownTags, WithTag("order", "int"), WithoutUnknownTags))
	todoFile := di.Config().TodoFile
	todos := `+P1 T1
+P1 T2
+P1 T3
+P1 to be deleted`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editAttempts(
		editSteps(appendOnEach("order:pmax+10"), appendInvalidTag(3)),
		editSteps(removeLine(3), reverse()),
	))

	cmd, ctx := cmd.Root(di)
	cmd.SetArgs([]string{"edit", "-s", ""})
	cmd.SetIn(strings.NewReader("Y\n"))
	err := cmd.ExecuteContext(ctx)

	assert.Nil(t, err)
	lines := ReadLines(t, todoFile)
	assert.Len(t, lines, 3)
	assert.ElementsMatch(t, ReadLines(t, todoFile), []string{
		"+P1 T1 order:30",
		"+P1 T2 order:20",
		"+P1 T3 order:10",
	})
}

func Test_TagExpansionIsEvaluatedInOrderOfTheEditFileEvenEvenWhenThereAreNewItems(t *testing.T) {
	di := BuildTestDi(t, BuildTestConfig(t, WithoutUnknownTags, WithTag("order", "int")))
	todoFile := di.Config().TodoFile
	todos := `+P1 T1
+P1 T2`
	assert.Nil(t, os.WriteFile(todoFile, []byte(todos), 0644))

	di.SetEditor(editSteps(appendLine("+P1 T3"), appendOnEach("order:pmax+10")))

	cmd, ctx := cmd.Root(di)
	cmd.SetArgs([]string{"edit", "-s", ""})
	cmd.SetIn(strings.NewReader("Y\n"))
	err := cmd.ExecuteContext(ctx)

	assert.Nil(t, err)
	lines := ReadLines(t, todoFile)
	assert.Len(t, lines, 3)
	assert.ElementsMatch(t, ReadLines(t, todoFile), []string{
		"+P1 T1 order:10",
		"+P1 T2 order:20",
		"+P1 T3 order:30",
	})
}

func editAttempts(editors ...di.Editor) di.Editor {
	invocationCount := 0
	return di.EditorFunc(func(path string) error {
		err := editors[invocationCount].Edit(path)
		invocationCount++
		return err
	})
}

func editSteps(editors ...di.Editor) di.Editor {
	return di.EditorFunc(func(path string) error {
		for _, e := range editors {
			err := e.Edit(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func appendOnEach(suffix string) di.Editor {
	return di.EditorFunc(func(path string) error {
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

func appendLine(line string) di.Editor {
	return di.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines = append(lines, line)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func appendInvalidLine() di.Editor {
	return appendLine("2022-13-01 Hello World")
}

func appendInvalidTag(idx int) di.Editor {
	return di.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines[idx] = fmt.Sprintf("%s %s", lines[idx], "invalid:tag")
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func removeLine(idx int) di.Editor {
	return di.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines = slices.Delete(lines, idx, idx+1)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func complete(completeTask int) di.Editor {
	return di.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		lines[completeTask] = fmt.Sprintf("x %s", lines[completeTask])
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func reverse() di.Editor {
	return di.EditorFunc(func(path string) error {
		lines, err := ReadLinesE(path)
		if err != nil {
			return err
		}
		slices.Reverse(lines)
		return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	})
}

func shuffle() di.Editor {
	return di.EditorFunc(func(path string) error {
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
