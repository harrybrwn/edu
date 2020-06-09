package main

import (
	"github.com/harrybrwn/edu/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		cmd.Stop(err)
	}
}
