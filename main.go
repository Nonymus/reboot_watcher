package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
)

var (
	sentinelPath string
	promFile     string
	metric       string
)

//go:generate ./scripts/get_version.sh
//go:embed version.txt
var version string

func watchSentinel() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Can't watch on non-existant file, so watch the folder instead
	dir := path.Dir(sentinelPath)
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == sentinelPath {
				if event.Op&(fsnotify.Create|fsnotify.Write) > 0 {
					updateFile(1)
					log.Println(event)
				}
				if event.Op&(fsnotify.Remove|fsnotify.Rename) > 0 {
					updateFile(0)
					log.Println(event)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func updateFile(value int) {
	err := func() error {
		f, err := os.Create(promFile + ".tmp")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "%s %d\n", metric, value); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		if err := os.Rename(f.Name(), promFile); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		log.Println("error: ", err)
	}
}

func setupFlags() {
	flag.StringVar(&sentinelPath, "sentinel", "/var/run/reboot-required", "path to sentinel file")
	flag.StringVar(&promFile, "promfile", "/var/lib/node_exporter/reboot.prom", "path to promfile")
	flag.StringVar(&metric, "metric", "node_reboot_required", "Prometheus metric name")
	flag.Parse()
}

func work() {
	log.Println("Version:", strings.TrimSpace(version))
	log.Println("sentinel:", sentinelPath)
	log.Println("promfile:", promFile)
	// Initial check
	if _, err := os.Stat(sentinelPath); err == nil {
		updateFile(1)
	} else if errors.Is(err, os.ErrNotExist) {
		updateFile(0)
	} else {
		log.Fatal(err)
	}
	// Watch dir for file creation/deletion
	go watchSentinel()
	// Wait till killed
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}

func main() {
	setupFlags()
	work()
}
