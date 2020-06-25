package term

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
)

var (
	// Output is the default output for the terminal package
	Output io.Writer

	escape string
)

const (
	FgBlack = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

const (
	Reset = iota
	Bold
	Faint
	Italic
	Underlined
	BlinkSlow
	BlinkFast
	Inverted
	Hidden
	CrossedOut
)

func init() {
	switch runtime.GOOS {
	case "windows":
		Output = ioutil.Discard
		escape = ""
	default:
		Output = os.Stdout
		escape = "\x1b"
	}
}

// Color256 outputs s in the 256 color format.
func Color256(code uint8, s string) string {
	return fmt.Sprintf("%s[38;5;%dm%s%s[0m", escape, code, s, escape)
}

// CursorOn turns the cursor on
func CursorOn() { out(control("?25h")) }

// CursorOff turns the cursor off
func CursorOff() { out(control("?25l")) }

// Blue returns s but colored blue
func Blue(s string) string { return color(36, s) }

// Green returns s but colored green
func Green(s string) string { return color(FgGreen, s) }

// Red returns s but colored red
func Red(s string) string { return color(FgRed, s) }

// Yellow returns s but colored yellow
func Yellow(s string) string { return color(FgYellow, s) }

func color(color int, s string) string {
	if escape == "" {
		return s
	}
	return fmt.Sprintf("%[1]s[%dm%s%[1]s[0m", escape, color, s)
}

func out(s string) {
	fmt.Fprint(Output, s)
}

func control(code string) string {
	if escape == "" {
		return ""
	}
	return fmt.Sprintf("%s[%s", escape, code)
}
