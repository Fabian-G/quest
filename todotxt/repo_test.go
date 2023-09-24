package todotxt_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_ItemsAreOrderedCorrectly(t *testing.T) {
	file := createTestFile(t, `x item 1
	item 3
	item 2`)
	repo := todotxt.NewRepo(file)
	repo.DefaultOrder = func(i1, i2 *todotxt.Item) int {
		return -1 * strings.Compare(i1.Description(), i2.Description())
	}

	list, err := repo.Read()

	assert.Nil(t, err)
	assert.Equal(t, 3, list.Len())
	assert.Equal(t, "item 3", list.Get(1).Description())
	assert.Equal(t, "item 2", list.Get(2).Description())
	assert.Equal(t, "item 1", list.Get(3).Description())
}

func Test_OptimisticLockingReturnsErrorOnSaveIfWrittenInTheMeantime(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	appendNewTask(t, file, "A new todo item")
	assert.Nil(t, err)
	err = repo.Save(list)

	var oError todotxt.OLockError
	assert.ErrorAs(t, err, &oError)
	content, err := os.ReadFile(oError.BackupPath)
	assert.Nil(t, err)
	assert.Equal(t, "A todo item\n", string(content))
}

func Test_OptimisticLockingDoesNotReturnErrorIfFileWasNotChanged(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	list, err := repo.Read()
	assert.Nil(t, err)
	err = repo.Save(list)
	assert.Nil(t, err)
}

func Test_OptimisticLockingIsNotTriggeredOnDoubleSave(t *testing.T) {
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	err := repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("Hello World"))))
	assert.Nil(t, err)
	err = repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("Hello World2"))))
	assert.Nil(t, err)
}

func Test_SaveTruncatesTheFileCorrectly(t *testing.T) {
	file := createTestFile(t, `an item with a very long description`)
	repo := todotxt.NewRepo(file)

	err := repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("short desc"))))
	assert.Nil(t, err)

	content, err := os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, "short desc\n", string(content))
}

func Test_WatchSendsNotificationOnFileChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	file := createTestFile(t, `A todo item`)
	repo := todotxt.NewRepo(file)

	c, rm, err := repo.Watch()
	defer repo.Close()
	defer rm()
	assert.Nil(t, err)

	go triggerChange(t, file)
	timeout := time.After(15 * time.Second)
	var change todotxt.ReadFunc
	select {
	case change = <-c:
	case <-timeout:
		t.Fatal("file change was not detected in time")
	}
	newList, err := change()
	assert.Nil(t, err)
	assert.Equal(t, 1, newList.Len())
	assert.Equal(t, "Hello World", newList.Get(1).Description())
}

func triggerChange(t *testing.T, file string) {
	tmp, err := os.CreateTemp("", "quest-repo-test.*.part")
	assert.Nil(t, err)
	io.Copy(tmp, strings.NewReader("Hello World"))
	err = os.Rename(tmp.Name(), file)
	assert.Nil(t, err)
}

func Test_BackupsAreKeptAppropriately(t *testing.T) {
	file := createTestFile(t, "original item\n")
	extension := path.Ext(file)
	fileName := strings.TrimSuffix(path.Base(file), extension)
	backupName := func(n int) string {
		return fmt.Sprintf(".%s.quest-backup-%d%s", fileName, n, extension)
	}

	repo := todotxt.NewRepo(file)
	repo.Keep = 2

	err := repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("new item"))))
	assert.Nil(t, err)
	backupContent, err := os.ReadFile(path.Join(path.Dir(file), backupName(0)))
	assert.Nil(t, err)
	assert.Equal(t, "original item\n", string(backupContent))

	err = repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("another item"))))
	assert.Nil(t, err)
	backupContent, err = os.ReadFile(path.Join(path.Dir(file), backupName(1)))
	assert.Nil(t, err)
	assert.Equal(t, "original item\n", string(backupContent))
	backupContent, err = os.ReadFile(path.Join(path.Dir(file), backupName(0)))
	assert.Nil(t, err)
	assert.Equal(t, "new item\n", string(backupContent))

	err = repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("a last item"))))
	assert.Nil(t, err)
	_, err = os.Stat(path.Join(path.Dir(file), backupName(2)))
	assert.ErrorIs(t, err, fs.ErrNotExist)
	backupContent, err = os.ReadFile(path.Join(path.Dir(file), backupName(1)))
	assert.Nil(t, err)
	assert.Equal(t, "new item\n", string(backupContent))
	backupContent, err = os.ReadFile(path.Join(path.Dir(file), backupName(0)))
	assert.Nil(t, err)
	assert.Equal(t, "another item\n", string(backupContent))
}

func Test_NonExistingTodoFileIsCreated(t *testing.T) {
	tmpDir := createTmpDir(t)
	repo := todotxt.NewRepo(path.Join(tmpDir, "mytodo.txt"))

	err := repo.Save(todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("Hello World"))))
	assert.Nil(t, err)
	list, err := repo.Read()

	assert.Nil(t, err)
	assert.Equal(t, 1, list.Len())
	assert.Equal(t, "Hello World", list.Get(1).Description())
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
