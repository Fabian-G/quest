package todotxt

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type OLockError struct {
	BackupPath string
}

func (o OLockError) Error() string {
	return fmt.Sprintf("the file was changed since the last time we read it. Wrote to %s instead", o.BackupPath)
}

type ReadFunc func() (*List, error)

type Repo struct {
	file         string
	checksum     [20]byte
	watcher      *fsnotify.Watcher
	updateChan   []chan ReadFunc
	fileLock     sync.Mutex
	watchLock    sync.Mutex
	Encoder      *Encoder
	Decoder      *Decoder
	DefaultHooks []HookBuilder
	DefaultOrder func(*Item, *Item) int
	Keep         int
}

func NewRepo(dest string) *Repo {
	return &Repo{
		file: dest,
	}
}

func (t *Repo) Save(l *List) error {
	t.fileLock.Lock()
	defer t.fileLock.Unlock()
	err := t.handleOptimisticLocking(l)
	if err != nil {
		return fmt.Errorf("could not save file %s: %w", t.file, err)
	}
	err = t.write(l)
	if err != nil {
		return fmt.Errorf("could not save todo list: %w", err)
	}
	return nil
}

func (t *Repo) handleOptimisticLocking(l *List) error {
	if t.checksum == [20]byte{} {
		return nil // This is a save without a prior read, so we don't need locking
	}
	currentData, err := t.load()
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not determine checksum of current state")
	}
	if sha1.Sum(currentData) != t.checksum {
		return fmt.Errorf("locking error: %w", t.writeToAlternativeLocation(l))
	}
	return nil
}

func (t *Repo) writeToAlternativeLocation(l *List) error {
	extension := path.Ext(t.file)
	fileName := strings.TrimSuffix(path.Base(t.file), extension)
	tmp, err := os.CreateTemp(path.Dir(t.file), fmt.Sprintf("%s.quest-conflict-*%s", fileName, extension))
	if err != nil {
		return fmt.Errorf("could not write file to alternative location")
	}
	defer func() {
		tmp.Close()
	}()
	err = t.encoder().Encode(tmp, l.Tasks())
	if err != nil {
		return err
	}
	return OLockError{
		BackupPath: tmp.Name(),
	}
}

func (t *Repo) write(l *List) error {
	err := t.backup()
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	tmp, err := os.CreateTemp(path.Dir(t.file), ".quest.part.*")
	if err != nil {
		return fmt.Errorf("could not create temporary file: %w", err)
	}

	buffer := bytes.Buffer{}
	err = t.encoder().Encode(io.MultiWriter(tmp, &buffer), l.Tasks())
	if err != nil {
		return fmt.Errorf("could not write txt file %s: %w", t.file, err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("could not close tmp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), t.file); err != nil {
		return fmt.Errorf("could not move tmp file to final location: %w", err)
	}
	t.checksum = sha1.Sum(buffer.Bytes())
	return nil
}

func (t *Repo) backup() error {
	extension := path.Ext(t.file)
	fileName := strings.TrimSuffix(path.Base(t.file), extension)
	nameTemplate := fmt.Sprintf(".%s.quest-backup-%%d%s", fileName, extension)
	if t.Keep <= 0 {
		return nil
	}
	if err := os.Remove(path.Join(path.Dir(t.file), fmt.Sprintf(nameTemplate, t.Keep-1))); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for i := t.Keep - 2; i >= 0; i-- {
		origin := path.Join(path.Dir(t.file), fmt.Sprintf(nameTemplate, i))
		dest := path.Join(path.Dir(t.file), fmt.Sprintf(nameTemplate, i+1))
		if err := os.Rename(origin, dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	data, err := os.ReadFile(t.file)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(path.Dir(t.file), fmt.Sprintf(nameTemplate, 0)), data, 0644); err != nil {
		return err
	}
	return nil
}

func (t *Repo) Read() (*List, error) {
	t.fileLock.Lock()
	defer t.fileLock.Unlock()
	rawData, err := t.load()
	if err != nil {
		return nil, err
	}
	tasks, err := t.decoder().Decode(bytes.NewReader(rawData))
	if err != nil {
		return nil, fmt.Errorf("could not parse txt file %s: %w", t.file, err)
	}
	list := ListOf(tasks...)
	list.IdxOrderFunc = t.DefaultOrder
	list.Reindex()
	t.checksum = sha1.Sum(rawData)
	for _, b := range t.DefaultHooks {
		list.AddHook(b(list))
	}
	return list, list.validate()
}

func (t *Repo) load() ([]byte, error) {
	file, err := os.Open(t.file)
	if err != nil {
		return nil, fmt.Errorf("could not open txt file %s for reading: %w", t.file, err)
	}
	defer file.Close()

	rawData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read txt file %s: %w", t.file, err)
	}
	return rawData, nil
}

// Watch watches watches for file changes.
// The second return argument can be used to unregister the watcher.
// However, that function must not be called from the same go routine that listens on the
// returned channel. When in doubt you can close it asynchronosuly `go remove()`
func (t *Repo) Watch() (<-chan ReadFunc, func(), error) {
	t.watchLock.Lock()
	defer t.watchLock.Unlock()
	if t.watcher == nil {
		var err error
		t.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, nil, fmt.Errorf("could not create watcher: %w", err)
		}
		err = t.watcher.Add(t.file)
		if err != nil {
			return nil, nil, fmt.Errorf("could not start watching %s: %w", t.file, err)
		}
		go t.fileWatcher()
	}
	newChan := make(chan ReadFunc)
	t.updateChan = append(t.updateChan, newChan)
	remove := func() {
		t.watchLock.Lock()
		defer t.watchLock.Unlock()
		t.updateChan = slices.DeleteFunc(t.updateChan, func(i chan ReadFunc) bool { return i == newChan })
		close(newChan)
	}
	return newChan, remove, nil
}

func (t *Repo) fileWatcher() {
	defer func() {
		t.watcher.Close()
		t.watcher = nil
	}()
	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Chmod) {
				t.notify()
			}
			if event.Has(fsnotify.Remove) {
				// try re-adding the file. If that does not work we are lost.
				err := t.watcher.Add(t.file)
				if err != nil {
					return
				}
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				return
			}
			log.Println(err)
		}
	}
}

func (t *Repo) notify() {
	t.watchLock.Lock()
	defer t.watchLock.Unlock()
	for _, c := range t.updateChan {
		c <- t.Read
	}
}

func (t *Repo) Close() {
	t.watchLock.Lock()
	defer t.watchLock.Unlock()
	if t.watcher == nil {
		return
	}
	t.watcher.Close()
	for _, c := range t.updateChan {
		close(c)
	}
	t.updateChan = nil
}

func (t *Repo) encoder() *Encoder {
	if t.Encoder != nil {
		return t.Encoder
	}
	return &DefaultEncoder
}

func (t *Repo) decoder() *Decoder {
	if t.Decoder != nil {
		return t.Decoder
	}
	return &DefaultDecoder
}
