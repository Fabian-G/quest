package di

import (
	"github.com/Fabian-G/quest/todotxt"
)

func buildNotesRepo(c Config) *todotxt.NotesRepo {
	if len(c.Notes.Tag) == 0 {
		return nil
	}
	repo := todotxt.NewNotesRepo(c.Notes.Tag, c.Notes.Dir)
	repo.IdLength = c.Notes.IdLength
	return repo
}
