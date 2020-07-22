package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestFiles(t *testing.T) {
	fmt.Println(os.Getenv("USER"))
	c := exec.Command("systemctl", "status", "edu")
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%T\n", err)
	if e, ok := err.(*exec.ExitError); ok {
		fmt.Println(e.ExitCode())
	}
}
