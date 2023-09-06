package main

import (
	"flag"
	"log"
	"os"
	"time"
)

// Wait for a path to exist

var (
	l     = log.New(os.Stderr, "🟧️  wait-fs: ", log.Ldate|log.Ltime|log.Lshortfile)
	path  string
	wait  int = 5
	quiet bool
)

// exists determines when a path exists (returning true)
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func main() {

	flag.StringVar(&path, "p", "", "path to look for")
	flag.IntVar(&wait, "w", 5, "seconds to wait (defaults to 5)")
	flag.BoolVar(&quiet, "q", false, "quiet level messaging")
	flag.Parse()
	timeout := time.Duration(wait)

	if path == "" {
		l.Fatal("A path (-p) is required")
	}
	if !quiet {
		l.Printf("%s\n", path)
	}

	// Stay in loop until the path exists
	var exists bool
	for {
		exists = fileExists(path)
		if exists {
			if !quiet {
				l.Printf("Found existing path %s\n", path)
			}
			break
		}
		l.Printf("Path %s does not exist yet, sleeping %d\n", path, timeout)
		time.Sleep(timeout * time.Second)
	}
}
