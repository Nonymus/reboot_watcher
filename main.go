package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	sentinelPath        string
	listen              string
	promFile            string
	rebootRequiredGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "node",
		Name:      "reboot_required",
		Help:      "OS requires reboot",
	})
)

func rebootRequired() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

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
					updateFile(true)
					rebootRequiredGauge.Set(1)
					log.Println(event)
				}
				if event.Op&(fsnotify.Remove|fsnotify.Rename) > 0 {
					updateFile(false)
					rebootRequiredGauge.Set(0)
					log.Println(event)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error: ", err)
		}
	}
}

func updateFile(value bool) {
	var code int
	if value {
		code = 1
	} else {
		code = 0
	}
	f, err := os.Create(promFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "node_reboot_required %d", code); err != nil {
		log.Fatal(err)
	}
}

func init() {
	prometheus.MustRegister(rebootRequiredGauge)
	flag.StringVar(&sentinelPath, "sentinelPath", "/var/run/reboot-required", "path to sentinel file")
	flag.StringVar(&listen, "listen", ":8080", "listen string (IP:port)")
	flag.StringVar(&promFile, "promFilePath", "/var/lib/node_exporter/reboot.prom", "path to promfile")
	flag.Parse()
}

func main() {
	log.Println("Sentinel: ", sentinelPath)
	// Initial check
	if _, err := os.Stat(sentinelPath); err == nil {
		rebootRequiredGauge.Set(1)
	} else if errors.Is(err, os.ErrNotExist) {
		rebootRequiredGauge.Set(0)
	} else {
		log.Fatal(err)
	}
	// Watch dir for file creation/deletion
	go rebootRequired()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listen, nil))
}
