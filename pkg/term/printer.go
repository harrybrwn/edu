package term

import (
	"bytes"
	"fmt"
)

// Colorf will generate a formatted string based on color format codes,
// so basically fmt.Sprintf but for terminal colors.
//
// Syntax:
//	%[modifier]<forground>
//
// Forground Colors:
//	%r - red
//	%g - green
//	%y - yellow
//	%b - blue
//	%m - magenta
//	%c - cyan
//	%w - white
//	%0 - reset (no op)
//
// Modifiers:
//	%!r - red bold
//	%.r - red faint
//	%'r - red italic
//	%_r - red underlined
//	%?r - red inverted
//	% r - red hidden
//	%-r - red crossed out
func Colorf(format string, a ...interface{}) string {
	var (
		modifier int
		color    int
		end      = len(format)
		arg      = 0
		buf      = &bytes.Buffer{}
	)

	for i := 0; i < end; i++ {
		prev := i
		for i < end && format[i] != '%' {
			i++
		}
		if i > prev {
			buf.WriteString(format[prev:i])
		}
		if i >= end {
			break
		}

		i++
		chr := format[i]

		switch chr {
		case '%': // if %% is in the format
			buf.WriteByte('%')
			continue
		case '!':
			modifier = Bold
		case '.':
			modifier = Faint
		case '\'':
			modifier = Italic
		case '_':
			modifier = Underlined
		case '*':
			modifier = BlinkSlow
			if format[i+1] == '*' {
				modifier = BlinkFast
				i++
			}
		case '?':
			modifier = Inverted
		case ' ':
			modifier = Hidden
		case '-':
			modifier = CrossedOut
		}
		if modifier != 0 {
			i++
		}
		color = getForground(format[i])

		if modifier != 0 {
			fmt.Fprintf(buf, "%s[%d;%dm", escape, color, modifier)
		} else {
			fmt.Fprintf(buf, "%s[%dm", escape, color)
		}
		buf.WriteString(fmt.Sprintf("%v", a[arg]))
		fmt.Fprintf(buf, "%s[0m", escape)
		arg++
	}
	return buf.String()
}

// ColorPrintf is the same as Colorf except it prints the result
// to standar out.
func ColorPrintf(format string, v ...interface{}) {
	fmt.Printf(Colorf(format, v...))
}

func getForground(code byte) int {
	switch code {
	case 'r':
		return FgRed
	case 'g':
		return FgGreen
	case 'y':
		return FgYellow
	case 'b':
		return FgBlue
	case 'm':
		return FgMagenta
	case 'c':
		return FgCyan
	case 'w':
		return FgWhite
	case '0':
		return Reset
	}
	return Reset
}
