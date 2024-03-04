package todotxt_test

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

var testNotesTag = "quest-test-note"

func Test_NewNoteIsCreatedIfNoneExists(t *testing.T) {
	notesDir := createTmpDir(t)
	notesRepo := todotxt.NewNotesRepo(testNotesTag, notesDir)

	testItem := todotxt.MustBuildItem(todotxt.WithDescription("Test item"))

	note, err := notesRepo.Get(testItem)
	assert.NoError(t, err)

	noteTags, ok := testItem.Tags()[testNotesTag]
	assert.True(t, ok)
	assert.Len(t, noteTags, 1)
	assert.Equal(t, path.Join(notesDir, fmt.Sprintf("%s.md", noteTags[0])), note)
	noteFile, err := os.Open(note)
	assert.NoError(t, err)
	defer noteFile.Close()
	content, err := io.ReadAll(noteFile)
	assert.NoError(t, err)
	assert.Equal(t, "# Notes for task \"Test item\"\n", string(content))
}

func Test_ExistingNoteIsReturnedIfOneExists(t *testing.T) {
	notesDir := createTmpDir(t)
	notesRepo := todotxt.NewNotesRepo(testNotesTag, notesDir)

	testItem := todotxt.MustBuildItem(todotxt.WithDescription("Test item"))

	// Create note and append something
	note, err := notesRepo.Get(testItem)
	assert.NoError(t, err)
	noteFileToAppend, err := os.OpenFile(note, os.O_APPEND|os.O_RDWR, 0000)
	assert.NoError(t, err)
	_, err = noteFileToAppend.WriteString("Some more content")
	assert.NoError(t, err)
	noteFileToAppend.Close()
	noteFile, err := os.Open(note)
	assert.NoError(t, err)
	defer noteFile.Close()
	content, err := io.ReadAll(noteFile)

	// Get the same note again
	note2, err := notesRepo.Get(testItem)
	assert.NoError(t, err)
	noteFile2, err := os.Open(note2)
	assert.NoError(t, err)
	defer noteFile2.Close()
	content2, err := io.ReadAll(noteFile2)

	assert.Equal(t, note, note2)
	assert.Equal(t, content, content2)
}

func Test_AnErrorIsReturnedWhenRunningOutOfIds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}

	notesDir := createTmpDir(t)
	for i := 0; i < 10; i++ {
		file, err := os.Create(path.Join(notesDir, fmt.Sprintf("%d.md", i)))
		assert.NoError(t, err)
		file.Close()
	}
	for i := 0; i < 26; i++ {
		file, err := os.Create(path.Join(notesDir, fmt.Sprintf("%s.md", string(rune('a'+i)))))
		assert.NoError(t, err)
		file.Close()
	}
	notesRepo := todotxt.NewNotesRepo(testNotesTag, notesDir)
	notesRepo.IdLength = 1

	item := todotxt.MustBuildItem(todotxt.WithDescription("Test item"))

	_, err := notesRepo.Get(item)
	assert.ErrorIs(t, err, todotxt.NoNoteIdsError)
}

func Test_CleanRemovesUnreferencedNotes(t *testing.T) {
	notesDir := createTmpDir(t)
	notesRepo := todotxt.NewNotesRepo(testNotesTag, notesDir)

	testItem := todotxt.MustBuildItem(todotxt.WithDescription("Test item"))
	otherTestItem := todotxt.MustBuildItem(todotxt.WithDescription("Other Test item"))
	list := todotxt.ListOf(testItem, otherTestItem)

	note1, err := notesRepo.Get(testItem)
	assert.NoError(t, err)
	note2, err := notesRepo.Get(otherTestItem)
	assert.NoError(t, err)

	otherTestItem.SetTag(testNotesTag, "")
	notesRepo.Clean(list)

	stat, err := os.Stat(note1)
	assert.NoError(t, err)
	assert.True(t, stat.Mode().IsRegular())
	_, err = os.Stat(note2)
	assert.ErrorIs(t, err, os.ErrNotExist)

}
