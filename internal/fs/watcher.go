package fs

import (
	"context"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Op   uint8
	Path string
}

const (
	Create uint8 = 1 << iota
	Write
	Remove
	Rename
)

type Watcher struct {
	ctx context.Context
}

func NewWatcher(ctx context.Context) (*Watcher, error) {
	return &Watcher{ctx: ctx}, nil
}

func (w *Watcher) Watch(directories []string, events chan<- Event) error {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return errors.Wrap(err, "initiate new filesystem watcher")
	}

	defer func() {
		if err = watcher.Close(); err != nil {
			log.Errorf("close file watcher: %v", err)
		}
	}()

	for _, directory := range directories {
		err = watcher.Add(directory)

		if err != nil {
			log.Errorf("add directory to watch: %v", err)
		}
	}

loop:
	for {
		select {
		case <-w.ctx.Done():
			break loop
		case event, ok := <-watcher.Events:
			if !ok {
				break loop
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				fi, err := os.Stat(event.Name)

				if err == nil && fi != nil && fi.IsDir() {
					err = watcher.Add(event.Name)

					if err != nil {
						log.Errorf("add new subdirectory to watch: %v", err)
					}

					continue
				}

				events <- Event{
					Op:   Create,
					Path: event.Name,
				}
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				events <- Event{
					Op:   Write,
					Path: event.Name,
				}
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				// try to stop watching subdirectory
				// ignore errors since we can't know if it's directory or not
				_ = watcher.Remove(event.Name)

				events <- Event{
					Op:   Remove,
					Path: event.Name,
				}
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				events <- Event{
					Op:   Rename,
					Path: event.Name,
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break loop
			}

			log.Errorf("watcher error: %v", err)
		}
	}

	return nil
}
