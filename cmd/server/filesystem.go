package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"fleek-technical-test/internal/fs"
	"fleek-technical-test/internal/store"
)

func watchFileSystem(ctx context.Context, store store.Store, src, dst string) {
	// tracker map is used to keep an track on last file hash of a file path
	// this is useful to delete old encrypted file when original one is modified
	tracker := make(map[string]string)

	// directoriesToWatch list main source directory plus every subdirectories at startup to watch
	directoriesToWatch := []string{src}

	log.Debug("scanning source directory at startup")

	// at startup, we scan disk to create encrypted files
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// add subdirectory to watch list
			directoriesToWatch = append(directoriesToWatch, path)
			return nil
		}

		file, err := writeFile(path, src, dst, store)

		if err != nil {
			return errors.Wrap(err, "create file")
		}

		tracker[path] = file.Hash()

		log.WithFields(log.Fields{
			"path": file.Path(),
			"hash": file.Hash(),
		}).Info("file created")

		return nil
	})

	if err != nil {
		log.Errorf("scan source directory: %v", err)
	}

	log.Debug("start filesystem watcher")

	// and after we start filesystem watcher for new incoming events (create/update/delete)
	watcher, err := fs.NewWatcher(ctx)

	if err != nil {
		log.Fatal(err)
	}

	// we create buffered channel to be able to received multiple events without blocking fsnotify while we encrypt files
	events := make(chan fs.Event, 10)

	go func() {
		err = watcher.Watch(directoriesToWatch, events)

		if err != nil {
			log.Fatal(err)
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case event := <-events:
			switch event.Op {
			case fs.Create:
				// FIXME sleep 1 second to let filesystem finish to properly write file on disk
				// otherwise the encryption will be corrupted, need to find a better alternative
				time.Sleep(1*time.Second)

				file, err := writeFile(event.Path, src, dst, store)

				if err != nil {
					log.Println("create file:", err)
					continue
				}

				tracker[event.Path] = file.Hash()
				log.Println("file created:", file.Path(), file.Hash())
			case fs.Write:
				// FIXME same as create
				time.Sleep(1*time.Second)

				// retrieve previous hash of file
				previousHash, exist := tracker[event.Path]

				// we create new file
				file, err := writeFile(event.Path, src, dst, store)

				if err != nil {
					log.Println("update file:", err)
					continue
				}

				tracker[event.Path] = file.Hash()

				if exist {
					// and delete previous one
					store.Delete(previousHash)
					_ = os.Remove(dst + string(os.PathSeparator) + previousHash)
				} else {
					log.Println("trying to delete non-existent file or a directory")
				}

				log.Println("file updated:", file.Path(), file.Hash())
			case fs.Remove:
				// retrieve and delete previous hash of file
				previousHash, exist := tracker[event.Path]

				if !exist {
					log.Println("trying to delete non-existent file")
					continue
				}

				store.Delete(previousHash)
				_ = os.Remove(dst + string(os.PathSeparator) + previousHash)
				log.Println("file deleted:", event.Path, previousHash)
			}
		}
	}
}

func writeFile(path, src, dst string, store store.Store) (*fs.File, error) {
	input, err := os.Open(path)

	if err != nil {
		return nil, errors.Wrap(err, "open input file")
	}

	defer input.Close()

	fileInfo, err := os.Stat(path)

	if err != nil {
		return nil, errors.Wrap(err, "retrieve file info")
	}

	// clean filename, remove "./" and source directory path and first slash if present
	src = strings.TrimPrefix(src, "./")
	filename := strings.Replace(path, src, "", 1)
	filename = strings.TrimPrefix(filename, "/")

	file, err := fs.NewFile(path, input, filename, fileInfo.Size())

	if err != nil {
		return nil, errors.Wrap(err, "create new fs file")
	}

	// we need to return to the beginning of the file after calculating the sha-256 hash,
	// this is needed to be able to encrypt the file properly
	_, _ = input.Seek(0, io.SeekStart)

	output, err := os.Create(dst + string(os.PathSeparator) + file.Hash())

	if err != nil {
		return nil, errors.Wrap(err, "create output file")
	}

	defer output.Close()

	if err = fs.EncryptDecrypt(file.Key(), input, output); err != nil {
		return nil, errors.Wrap(err, "encrypt file")
	}

	// use SHA-256 hash of (filename + file content) as store key
	store.Add(file.Hash(), file)

	return file, nil
}
