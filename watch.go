package blueprint

import (
	"github.com/howeyc/fsnotify"
	"path/filepath"
	"regexp"
	"time"
)

var temp = regexp.MustCompile(`\.tmp$|\.TMP$`)

func BeginWatching() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	changes := make(chan string, 10)
	go worker(changes, "generated")

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if !temp.MatchString(ev.Name) {
					log.Println("event:", ev)
					if ev.IsCreate() || ev.IsModify() || ev.IsRename() {
						changes <- ev.Name
					}
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(bp.dir)
	if err != nil {
		log.Fatal(err)
	}
}

func worker(in <-chan string, gen string) {
	for name := range in {
		name = filepath.Base(name)
		time.Sleep(500 * time.Millisecond)

		if name == bp.master || isOneOf(name, bp.extra...) {
			if err := CompileDir(bp.master, bp.dir, bp.extra...); err != nil {
				log.Print("Recompile Dir Failed: ", err)
			} else {
				log.Printf("Recompiled Dir (%s)", name)
				SaveGenerated(gen)
			}
		} else {
			if err := CompileTemplate(name, bp.master, bp.dir, bp.extra...); err != nil {
				log.Printf("Compile failed (%s): %v", name, err)
			} else {
				log.Printf("Recompiled %s", name)
				SaveTemplate(name, gen)
			}
		}
	}
}
