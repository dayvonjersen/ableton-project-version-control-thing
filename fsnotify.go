package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type event struct {
	filename string
	t        time.Time
}

type watcher struct {
	w         *fsnotify.Watcher
	paths     []string
	events    chan *event
	validator func(string) bool
	callback  func(string)
}

func newWatcher(validator func(string) bool, callback func(string)) (*watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watch := &watcher{
		w:         w,
		paths:     []string{},
		events:    make(chan *event),
		validator: validator,
		callback:  callback,
	}

	go watch.relay()
	go watch.dispatch()

	return watch, nil
}

func (w *watcher) relay() {
	for {
		select {
		case e := <-w.w.Events:
			w.events <- &event{
				filename: normalizePathSeparators(e.Name),
				t:        time.Now(),
			}
		case err := <-w.w.Errors:
			checkErr(err)
		}
	}
}

func (w *watcher) dispatch() {
	var last time.Time
	for e := range w.events {
		// log.Println("dispatcher got:", path.Base(e.filename), diff)
		if !w.validator(e.filename) {
			// log.Println("file is not valid,          skipping...")
			continue
		}
		diff := time.Since(last) - time.Since(e.t)
		if diff < time.Millisecond*100 {
			// log.Println("last event was < 100ms ago, skipping...")
			continue
		}
		last = e.t
		go w.callback(e.filename)
	}
}

func (w *watcher) AddWithSubdirs(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, ".git") {
			path = normalizePathSeparators(path)
			// log.Println("watching", path)
			w.paths = append(w.paths, path)
			checkErr(w.w.Add(path))
		}
		return err
	})
}

func (w *watcher) RemoveAll() {
	for _, path := range w.paths {
		if err := w.w.Remove(path); err != nil {
			log.Println(err)
		}
	}
	w.paths = []string{}
}
