package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
)

func watchFolder(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if isWatchedFile(ev.Name) {
					watcherLog("sending event %s", ev)
					startChannel <- ev.String()
				}
			case err := <-watcher.Error:
				watcherLog("error: %s", err)
			}
		}
	}()

	watcherLog("Watching %s", path)
	err = watcher.Watch(path)

	if err != nil {
		fatal(err)
	}
}

func watch() {
	root := root()
	watcherLog("Start root: %v", root)
	scanLower := scanLower()
	if scanLower > 0 {
		wtch, err := filepath.Abs(root)
		if err == nil {
			root = wtch
		}
		split := strings.Split(strings.TrimRight(root, "/"), "/")
		if scanLower < len(split) {
			root = strings.Join(split[0:len(split)-scanLower], "/")
		}
	}
	watcherLog("Modified root: %v", root)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !isTmpDir(path) {
			if len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}

			if isIgnoredFolder(path) {
				watcherLog("Ignoring %s", path)
				return filepath.SkipDir
			}

			watchFolder(path)
		}

		return err
	})
}
