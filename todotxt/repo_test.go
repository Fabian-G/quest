package todotxt_test

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_OptimisticLockingReturnsErrorOnSaveIfWrittenInTheMeantime(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	appendNewTask(t, file, "A new todo item")
	assert.Nil(t, err)
	err = repo.Save(list)

	assert.ErrorIs(t, err, todotxt.ErrOLocked)
}

func Test_OptimisticLockingDoesNotReturnErrorIfFileWasNotChanged(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	err = repo.Save(list)

	assert.Nil(t, err)
}

func Test_WatchSendsNotificationOnFileChanges(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	c, rm, err := repo.Watch()
	defer repo.Close()
	defer rm()
	assert.Nil(t, err)

	go func() {
		f, err := os.OpenFile(file, os.O_RDWR, 0644)
		assert.Nil(t, err)
		defer f.Close()
		io.Copy(f, strings.NewReader("Hello World"))
	}()

	change := <-c
	newList, err := change()
	assert.Nil(t, err)
	assert.Len(t, newList, 1)
	assert.Equal(t, "Hello World", newList[0].Description())
}

func Test_NonExistingTodoFileIsCreated(t *testing.T) {
	tmpDir := createTmpDir(t)
	repo := todotxt.NewRepo(path.Join(tmpDir, "mytodo.txt"))

	err := repo.Save(todotxt.List{todotxt.MustBuildItem(todotxt.WithDescription("Hello World"))})
	assert.Nil(t, err)
	list, err := repo.Read()

	assert.Nil(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Hello World", list[0].Description())
}

func createTestFile(t *testing.T, content string) string {
	p := createTmpDir(t)

	f, err := os.OpenFile(path.Join(p, "todo.txt"), os.O_CREATE|os.O_RDWR, 0644)
	assert.Nil(t, err)
	defer f.Close()
	io.Copy(f, strings.NewReader(content))
	return f.Name()
}

func createTmpDir(t *testing.T) string {
	p, err := os.MkdirTemp("", "txtrepotest_*")
	assert.Nil(t, err)
	t.Cleanup(func() {
		err := os.RemoveAll(p)
		assert.Nil(t, err)
	})
	return p
}

func appendNewTask(t *testing.T, file string, new string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_RDWR, 0644)
	assert.Nil(t, err)
	defer f.Close()
	f.WriteString("\n")
	io.Copy(f, strings.NewReader(new))
	f.WriteString("\n")
}
