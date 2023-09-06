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
	"slices"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var ErrOLocked = errors.New("the file was changed since the last time we read it")

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
}

func NewRepo(dest string) *Repo {
	return &Repo{
		file: dest,
	}
}

func (t *Repo) Save(l *List) error {
	t.fileLock.Lock()
	defer t.fileLock.Unlock()
	err := t.handleOptimisticLocking()
	if err != nil {
		return fmt.Errorf("could not save file %s: %w", t.file, err)
	}
	err = t.write(l)
	if err != nil {
		return err
	}
	return nil
}

func (t *Repo) handleOptimisticLocking() error {
	currentData, err := t.load()
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not determine checksum of current state")
	}
	if sha1.Sum(currentData) != t.checksum {
		return ErrOLocked
	}
	return nil
}

func (t *Repo) write(l *List) error {
	file, err := os.OpenFile(t.file, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open txt file %s for writing: %w", t.file, err)
	}
	defer file.Close()

	tasksToWrite := l.Tasks()
	if t.DefaultOrder != nil {
		slices.SortStableFunc(tasksToWrite, t.DefaultOrder)
	}

	buffer := bytes.Buffer{}
	err = t.encoder().Encode(io.MultiWriter(file, &buffer), tasksToWrite)
	if err != nil {
		return fmt.Errorf("could not write txt file %s: %w", t.file, err)
	}
	t.checksum = sha1.Sum(buffer.Bytes())
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
	if t.DefaultOrder != nil {
		slices.SortStableFunc(tasks, t.DefaultOrder)
	}
	list := ListOf(tasks...)
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
			if event.Has(fsnotify.Write) {
				t.notify()
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
	t.watcher.Close()
	t.updateChan = nil
	for _, c := range t.updateChan {
		close(c)
	}
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
