package term

import (
	"fmt"
	"testing"
)

func TestPrint(t *testing.T) {
	s := Colorf("%s %r %b is blue %w two", "red", "blue", "cyan")
	fmt.Printf(s, "test")
}
