package disk

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/fsnotify/fsnotify"
)

type ReadFunc func() (todotxt.List, error)

type TxtRepo struct {
	file       string
	lastRead   int64
	watcher    *fsnotify.Watcher
	updateChan []chan ReadFunc
	fileLock   sync.Mutex
	watchLock  sync.Mutex
	Encoder    *todotxt.Encoder
	Decoder    *todotxt.Decoder
}

func NewTxtRepo(dest string) *TxtRepo {
	return &TxtRepo{
		file: dest,
	}
}

func (t *TxtRepo) Save(l todotxt.List) error {
	t.fileLock.Lock()
	defer t.fileLock.Unlock()
	stats, err := os.Stat(t.file)
	if err != nil {
		return fmt.Errorf("could not stat txt file %s: %w", t.file, err)
	}
	if stats.ModTime().Unix() > t.lastRead {
		return fmt.Errorf("destination file %s was written after the last time we read", t.file)
	}
	err = t.write(l)
	if err != nil {
		return err
	}
	t.lastRead = stats.ModTime().Unix()
	return nil
}

func (t *TxtRepo) write(l todotxt.List) error {
	file, err := os.OpenFile(t.file, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open txt file %s for writing: %w", t.file, err)
	}
	defer file.Close()

	err = t.encoder().Encode(file, l)
	if err != nil {
		return fmt.Errorf("could not write txt file %s: %w", t.file, err)
	}
	return nil
}

func (t *TxtRepo) Read() (todotxt.List, error) {
	t.fileLock.Lock()
	defer t.fileLock.Unlock()
	stats, err := os.Stat(t.file)
	if err != nil {
		return nil, fmt.Errorf("could not stat txt file %s: %w", t.file, err)
	}
	list, err := t.load()
	if err != nil {
		return nil, err
	}
	t.lastRead = stats.ModTime().Unix()
	return list, nil
}

func (t *TxtRepo) load() (todotxt.List, error) {
	file, err := os.Open(t.file)
	if err != nil {
		return nil, fmt.Errorf("could not open txt file %s for reading: %w", t.file, err)
	}
	defer file.Close()

	l, err := t.decoder().Decode(file)
	if err != nil {
		return nil, fmt.Errorf("could not parse txt file %s: %w", t.file, err)
	}
	return l, nil
}

// Watch watches watches for file changes.
// The second return argument can be used to unregister the watcher.
// However, that function must not be called from the same go routine that listens on the
// returned channel. When in doubt you can close it asynchronosuly `go remove()`
func (t *TxtRepo) Watch() (<-chan ReadFunc, func(), error) {
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

func (t *TxtRepo) fileWatcher() {
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
			log.Fatal(err)
		}
	}
}

func (t *TxtRepo) notify() {
	t.watchLock.Lock()
	defer t.watchLock.Unlock()
	for _, c := range t.updateChan {
		c <- t.Read
	}
}

func (t *TxtRepo) Close() {
	t.watchLock.Lock()
	defer t.watchLock.Unlock()
	t.watcher.Close()
	t.updateChan = nil
	for _, c := range t.updateChan {
		close(c)
	}
	t.watcher = nil
}

func (t *TxtRepo) encoder() *todotxt.Encoder {
	if t.Encoder != nil {
		return t.Encoder
	}
	return &todotxt.DefaultEncoder
}

func (t *TxtRepo) decoder() *todotxt.Decoder {
	if t.Decoder != nil {
		return t.Decoder
	}
	return &todotxt.DefaultDecoder
}
