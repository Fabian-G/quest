package todotxt

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

var maxIdIterations = 1000

var NoNoteIdsError = errors.New("could not find a new id in a reasonable amount of time. Try running \"quest notes clean\" or consider increasting id length")

type NotesRepo struct {
	dir      string
	tag      string
	IdLength int
}

func NewNotesRepo(notesTag string, notesDir string) *NotesRepo {
	return &NotesRepo{
		dir:      notesDir,
		tag:      notesTag,
		IdLength: 4,
	}
}

func nAlphaNum(n int) string {
	result := strings.Builder{}
	for i := 0; i < n; i++ {
		r := rand.Intn(36)
		if r <= 9 {
			_, _ = result.WriteString(strconv.Itoa(r))
		} else {
			_, _ = result.WriteRune(rune('a' + (r - 10)))
		}
	}
	return result.String()
}

func (n *NotesRepo) idToPath(id string) string {
	return path.Join(n.dir, fmt.Sprintf("%s.md", id))
}

func (n *NotesRepo) nextId() (string, error) {
	for i := 0; i < maxIdIterations; i++ {
		id := nAlphaNum(n.IdLength)
		_, err := os.Stat(n.idToPath(id))
		if errors.Is(err, os.ErrNotExist) {
			return id, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", NoNoteIdsError
}

func (n *NotesRepo) Get(item *Item) (string, error) {
	noteTags, ok := item.Tags()[n.tag]
	var note string
	if !ok || len(noteTags) == 0 {
		newNote, err := n.nextId()
		if err != nil {
			return "", err
		}
		note = newNote
		if err := item.SetTag(n.tag, note); err != nil {
			return "", fmt.Errorf("could not set reference to note: %w", err)
		}
	} else {
		note = noteTags[0]
	}
	_, err := os.Stat(n.idToPath(note))
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(n.idToPath(note))
		if err != nil {
			return "", err
		}
		defer file.Close()
		_, err = file.WriteString(fmt.Sprintf("# Notes for task \"%s\"\n",
			item.CleanDescription(item.Projects(), item.Contexts(), item.Tags().Keys())),
		)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat file %s: %w", n.idToPath(note), err)
	}

	return n.idToPath(note), nil
}

func (n *NotesRepo) Clean(lists ...*List) error {
	files := os.DirFS(n.dir)
	notes, err := fs.ReadDir(files, ".")
	if err != nil {
		return fmt.Errorf("could not read notes directory: %w", err)
	}

	referenced := make([]string, 0)
	for _, list := range lists {
		for _, item := range list.Tasks() {
			noteTags, ok := item.Tags()[n.tag]
			if !ok || len(noteTags) == 0 {
				continue
			}
			referenced = append(referenced, noteTags...)
		}
	}
	unreferencedNotes := slices.DeleteFunc(notes, func(entry fs.DirEntry) bool {
		if !entry.Type().IsRegular() {
			return true
		}
		return slices.Contains(referenced, strings.TrimSuffix(entry.Name(), ".md"))
	})
	for _, unref := range unreferencedNotes {
		err := os.Remove(path.Join(n.dir, unref.Name()))
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}
