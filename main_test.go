package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestWork(t *testing.T) {
	// Setup
	sentinelDir := t.TempDir()
	sentinelPath = sentinelDir + "/reboot-required"
	promDir := t.TempDir()
	promFile = promDir + "/reboot.prom"
	metric = "my_metric_name{foo=\"bar\"}"

	// Start and settle
	go work()
	time.Sleep(time.Millisecond * 100)

	t.Run("initalAbsent", func(t *testing.T) {
		// value should be zero
		f, _ := os.ReadFile(promFile)
		if string(f) != fmt.Sprintf("%s 0", metric) {
			t.Error("metric is not zero")
		}
	})

	t.Run("switchHigh", func(t *testing.T) {
		// value should switch to one
		os.WriteFile(sentinelPath, []byte(""), 0666)
		time.Sleep(time.Millisecond * 100)
		f, _ := os.ReadFile(promFile)
		if string(f) != fmt.Sprintf("%s 1", metric) {
			t.Error("metric is not one")
		}
	})

	t.Run("switchLow", func(t *testing.T) {
		// value should switch to zero when file deleted
		os.Remove(sentinelPath)
		time.Sleep(time.Millisecond * 100)
		f, _ := os.ReadFile(promFile)
		if string(f) != fmt.Sprintf("%s 0", metric) {
			t.Error("metric did not switch back to zero")
		}
	})
}
