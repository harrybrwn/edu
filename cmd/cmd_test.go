package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestFiles(t *testing.T) {
	osfile, err := os.OpenFile("./testfile.txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if os.IsExist(err) {
		fmt.Println("file exists")
		return
	}
	if err != nil {
		t.Error(err)
	}
	osfile.WriteString("hello")
}
