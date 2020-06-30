package commands

import (
	"testing"
)

func TestParseHTML(t *testing.T) {
	data := `<p>this <a>is</a> a test</p><br><p>this is a test</p>`
	err := parseHTML(data)
	if err != nil {
		t.Error(err)
	}
	println()
}
