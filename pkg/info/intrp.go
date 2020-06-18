package info

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
)

// Intrp starts a little mini interpreter that gives runtime stats
func Intrp() {
	s := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for s.Scan() {
		line := strings.Split(s.Text(), " ")
		cmd := line[0]
		args := line[1:]
		switch cmd {
		case "help", "h":
			println("Commands")
			println("  procs      - get info on goroutines and processes")
			println("  exit       - stop the program")
			println("  mem <file> - write a heap profile to <file>")
		case "procs":
			fmt.Println("goroutines:", runtime.NumGoroutine())
			fmt.Println("maxprocs:", runtime.GOMAXPROCS(0))
		case "mem":
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "no file argument")
				break
			}
			memprofile(args[0])
		case "exit", "quit", "q":
			os.Exit(0)
		}
		fmt.Print("> ")
	}
}

func memprofile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	defer file.Close()
	err = pprof.WriteHeapProfile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not write heap profile:", err)
	}
}
