package main

import (
	"flag"
	"log"
	"os"
	"time"

	ps "github.com/shirou/gopsutil/process"
)

// Wait for a process by name.
// This is used as a helper to first wait for a process id
// based on an identifiers

var (
	l          = log.New(os.Stderr, "🟧️  wait: ", log.Ldate|log.Ltime|log.Lshortfile)
	command    string
	executable string
	wait       int = 5
)

func main() {

	flag.StringVar(&command, "c", "", "command to look for")
	flag.StringVar(&executable, "e", "", "executable to look for")
	flag.IntVar(&wait, "w", 5, "seconds to wait (defaults to 5)")
	flag.Parse()
	timeout := time.Duration(wait)

	if command == "" && executable == "" {
		l.Fatal("A command (-c) or executable (-e) is required")
	}
	l.Printf("%s\n", command)

	for {
		procs, err := ps.Processes()
		if err != nil {
			l.Fatalf("Could not list processes %s\n", err)
		}
		for _, proc := range procs {
			exe, err := proc.Exe()
			if err != nil {
				continue
			}
			cmdline, err := proc.Cmdline()
			if err != nil {
				l.Fatalf("Error getting commandline %s\n", err)
			}
			if executable != "" && executable == exe {
				l.Printf("Found matched executable %s with pid %s\n", exe, proc.Pid)
				return
			}
			if command != "" && command == cmdline {
				l.Printf("Found matched command %s with pid %s\n", command, proc.Pid)
				return
			}
		}
		time.Sleep(timeout * time.Second)
		l.Printf("looking for matching pid...")
	}
}
