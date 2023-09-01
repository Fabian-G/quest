package disk

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OptimisticLockingReturnsErrorOnSaveIfWrittenInTheMeantime(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := NewTxtRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	appendNewTask(t, file, "A new todo item")
	// err = os.Chtimes(file, time.Now().Add(1*time.Second), time.Now().Add(1*time.Second))
	assert.Nil(t, err)
	err = repo.Save(list)

	assert.Error(t, err)
}

func appendNewTask(t *testing.T, file string, new string) {
	f, err := os.OpenFile(file, os.O_APPEND, 0644)
	assert.Nil(t, err)
	defer f.Close()
	f.WriteString("\n")
	io.Copy(f, strings.NewReader(new))
	f.WriteString("\n")
}

func Test_OptimisticLockingDoesNotReturnErrorIfFileWasNotChanged(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := NewTxtRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	err = repo.Save(list)

	assert.Nil(t, err)
}

func Test_WatchSendsNotificationOnFileChanges(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := NewTxtRepo(file)

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

func createTestFile(t *testing.T, content string) string {
	p, err := os.MkdirTemp("", "txtrepotest_*")
	assert.Nil(t, err)
	t.Cleanup(func() {
		err := os.RemoveAll(p)
		assert.Nil(t, err)
	})

	f, err := os.OpenFile(path.Join(p, "todo.txt"), os.O_CREATE|os.O_RDWR, 0644)
	assert.Nil(t, err)
	defer f.Close()
	io.Copy(f, strings.NewReader(content))
	return f.Name()
}
