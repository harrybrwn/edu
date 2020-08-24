package term

import (
	"testing"
)

func Test(t *testing.T) {
	// s := Colorf("%^r %1  %-b is blue %?0 two", "red", "red", "blue", "cyan")
	// println(s)
	// println(Color256(87, "hello"))
}

func TestColorf(t *testing.T) {
	tests := []struct {
		fmt, exp string
	}{
		{"%r", "\x1b[31mtest\x1b[0m"},
		{"%g", "\x1b[32mtest\x1b[0m"},
		{"%y", "\x1b[33mtest\x1b[0m"},
		{"%b", "\x1b[34mtest\x1b[0m"},
		{"%m", "\x1b[35mtest\x1b[0m"},
		{"%c", "\x1b[36mtest\x1b[0m"},
		{"%w", "\x1b[37mtest\x1b[0m"},

		{"%!r", "\x1b[31;1mtest\x1b[0m"},
		{"%.r", "\x1b[31;2mtest\x1b[0m"},
		{"%'r", "\x1b[31;3mtest\x1b[0m"},
		{"%_r", "\x1b[31;4mtest\x1b[0m"},
		{"%*r", "\x1b[31;5mtest\x1b[0m"},
		{"%**r", "\x1b[31;6mtest\x1b[0m"},
		{"%?r", "\x1b[31;7mtest\x1b[0m"},
		{"% r", "\x1b[31;8mtest\x1b[0m"},
		{"%-r", "\x1b[31;9mtest\x1b[0m"},

		// {"%1", "\x1b[38;5;1mtest\x1b[0m"},
	}
	for _, tst := range tests {
		res := Colorf(tst.fmt, "test")
		if res != tst.exp {
			t.Errorf("\x1b[0mwrong output for %s: got %s; want %s", tst.fmt, res, tst.exp)
		}
	}
}
